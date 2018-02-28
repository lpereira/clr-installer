#!/bin/bash -l

# This script is responsible for generating the necessary
# Travis-CI run script and executing the resulting script.
# This is done here instead on in the Groovy Jenkinsfile
# because we need to run this as the 'travis' user and trying
# to map all of these commands to work with 'sudo -u travis'
# was to complex and would loose the JENKINS environment.

rvm use 2.3.0

# Compile the .travis.yml into a Travis-CI script
cd ${WORKSPACE}
/home/travis/.travis/travis-build/bin/travis compile > ${HOME}/travis-ci.sh
