#!/bin/bash -l

# This script is responsible for generating the necessary
# Travis-CI run script and executing the resulting script.
# This is done here instead on in the Groovy Jenkinsfile
# because we need to run this as the 'travis' user and trying
# to map all of these commands to work with 'sudo -u travis'
# was to complex and would loose the JENKINS environment.

rvm use 2.3.0

# Ensure we are using the correct version of Go Lang
go_current=$(gimme -l |& grep current | awk '{print $1}')
eval $(/bin/bash ${WORKSPACE}/.Jenkins/yaml.sh)
if [ ${go_} != ${go_current} ]
then
    gimme ${go_} > ${HOME}/go_ver.sh
    echo -n "${go_}" > $HOME/.gimme/version
    . ${HOME}/go_ver.sh
    rm ${HOME}/go_ver.sh
fi

# Compile the .travis.yml into a Travis-CI script
cd ${WORKSPACE}
/home/travis/.travis/travis-build/bin/travis compile > ${HOME}/travis-ci.sh
