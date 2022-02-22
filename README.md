# Git proc-receive hook example

This example is implementation of Git
[proc-receive](https://git-scm.com/docs/githooks#proc-receive)
hook. In this example implementation it's trying to simulate
standard Git's receive-pack protocol and just updates references
from old OID to new one.
To test this example you need Git version >= 2.29.

Test it with:
 - `make all` - (default) build and test hook
 - `make build` - build hook's binary
 - `make prepare-test` - create test repositories
 - `make test-deploy` - deploy binary to test server repo
 - `make test` - run tests (see `Makefile` sources)
 - `make clean` - remove hook binary
 - `make test-clean` - remove test repos
