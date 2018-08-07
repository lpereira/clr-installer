#!/bin/bash

IMG=${IMG:-"clr-installer.img"}
CONF="clr-installer.yaml"

export EXTRA_BUNDLES=${EXTRA_BUNDLES:-""}

if [ -z "$CLR_INSTALLER_ROOT_DIR" ]; then
    SRCDIR=$(dirname $0)
    if [ -z "$SRCDIR" ]; then
        SRCDIR="."
    fi
    CLR_INSTALLER_ROOT_DIR=$(cd ${SRCDIR}/.. ; echo $PWD)
fi
echo "Using CLR_INSTALLER_ROOT_DIR=${CLR_INSTALLER_ROOT_DIR}"

INST_MAIN_DIR=$CLR_INSTALLER_ROOT_DIR/clr-installer
INST_TOML_FILE=$CLR_INSTALLER_ROOT_DIR/Gopkg.toml
INST_MAKEFILE=$CLR_INSTALLER_ROOT_DIR/Makefile

if [ ! -d $INST_MAIN_DIR ] || [ ! -f $INST_TOML_FILE ] || [ ! -f $INST_MAKEFILE ]; then
    echo "CLR_INSTALLER_ROOT_DIR doesn't point to the Clear Linux" \
        "OS Installer source dir"
    exit 1
fi

WORK_DIR=$PWD
echo "Creating empty image file {$IMG} ..."
rm -f {$WORK_DIR}/${IMG}
/usr/bin/qemu-img create -f raw ${WORK_DIR}/${IMG} 4G

echo "Enabling ${IMG} file for loopback..."
TEMP=$(mktemp -d)
LOOP=$(sudo losetup --find --show ${WORK_DIR}/${IMG})
LOOPNAME=$(basename ${LOOP})
DEVMAJMIN=$(losetup -O "MAJ:MIN" -n ${LOOP} | awk '{$1=$1};1')

echo "Using Loopback device ${LOOP} [${DEVMAJMIN}] ..." 

sudo partprobe $LOOP
sleep 2

echo "Creating installation configuration file ..."
sed -e "s/loop0/${LOOPNAME}/g;s/7:0/${DEVMAJMIN}/g" <<EOF >${WORK_DIR}/${CONF}
#clear-linux-config
targetMedia:
- name: loop0
  majMin: "7:0"
  size: "4294967296"
  ro: "false"
  rm: "false"
  type: loop
  children:
  - name: loop0p1
    fstype: vfat
    mountpoint: /boot
    size: "157286400"
    ro: "false"
    rm: "false"
    type: part
  - name: loop0p2
    fstype: swap
    size: "2147483648"
    ro: "false"
    rm: "false"
    type: part
  - name: loop0p3
    fstype: ext4
    mountpoint: /
    size: "1990197248"
    ro: "false"
    rm: "false"
    type: part
networkInterfaces: []
keyboard: us
language: en_US.UTF-8
bundles: [os-core, os-core-update, os-installer, telemetrics, ${EXTRA_BUNDLES}]
telemetry: false
timezone: America/Los_Angeles
kernel: kernel-native
postReboot: false
postArchive: false
autoUpdate: false
EOF

echo "Installing clear to loopback image file..."
pushd $CLR_INSTALLER_ROOT_DIR
sudo make
sudo -E $CLR_INSTALLER_ROOT_DIR/.gopath/bin/clr-installer --config ${WORK_DIR}/${CONF} --reboot=false
if [ $? -ne 0 ]
then
    echo "********************"
    echo "Install failed; Stopped image build process..."
    echo "********************"
    exit $?
fi

echo "Installing clr-installer into $TEMP"
sudo mount ${LOOP}p3 $TEMP

sudo make install DESTDIR=$TEMP
echo "Enabling clr-installer on boot for $TEMP"
sudo systemctl --root=$TEMP enable clr-installer

# Create a custom telemetry configuration to only log locally
echo "Creating custom telemetry configuration for $TEMP"
sudo /usr/bin/mkdir -p ${TEMP}/etc/telemetrics/
sudo /usr/bin/cp \
    ${TEMP}/usr/share/defaults/telemetrics/telemetrics.conf \
    ${TEMP}/etc/telemetrics/telemetrics.conf
sudo sed -i -e '/server=/s/clr.telemetry.intel.com/localhost/' \
    -e '/spool_process_time/s/=900/=3600/' \
    -e '/record_retention_enabled/s/=false/=true/' \
    ${TEMP}/etc/telemetrics/telemetrics.conf
popd

sudo umount $TEMP
sudo losetup -d $LOOP

exit 0
