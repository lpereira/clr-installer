# Clear Linux Installer

## Dependencies
The following bundles are required in order to run clr-installer:

+ sysadmin-basic (for kbd)
+ storage-utils
+ network-basic

## How to test?
Make sure you have any extra storage device, an USB memory stick should work fine, the installer will detect and use it if you choose.

## Clone this repository

```
git clone https://github.intel.com/iclr/clr-installer.git
```

## Build the installer

```
cd clr-installer && make
```

## Run as root

```
sudo ./bin/clr-installer
```

# Multiple Installer Modes
Currently the installer supports 2 modes (a third one is on the way):
1. Mass Installer - using an install descriptor file
2. TUI - a text based user interface
3. GUI - a graphical user interface (yet to come)

## Using Mass Installer
In order to use the Mass Installer provide a ```--config```, such as:

```
sudo ./bin/clr-installer --config ~/my-install.json
```

## Using TUI
Call the clr-installer executable without any additional flags, such as:

```
sudo ./bin/clr-installer
```

## Reboot
If you're running the installer on a development machine you may not want to reboot the system after the install completion, for that use the ```--reboot=false``` flag, such as:

```
sudo ./bin/clr-installer --reboot=false
```

or if using the Mass Installer mode:

```
sudo ./bin/clr-installer --config=~/my-install.json --reboot=false
```