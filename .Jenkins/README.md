# Enable Continuous Integration

# Travis-CI via Docker via Jenkins

For internal development, we still want to run the standard .travis.yml
CI test flow. We are using Jenkins to launch a pre-built Docker image
which has a local instance of Travis-CI.

See [Continuous Integration Environment](https://github.intel.com/iclr/ci-env)
repository for details on the creation and maintenance of the Docker image.
