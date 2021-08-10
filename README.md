# chainbing-node

Go implementation of the Chainbing node.

## Developing

To contribute to this codebase you must follow the [branch model and development flow](council/gitflow.md).

### Go version

The `node` has been tested with go version 1.14

### Build

- Build the binary in the local environment for you current OS and check the current version:
```shell
$ make
$ ./dist/cbnode version
```

- Build the binary in a docker container for all supported OS and check the current version (only docker needed):
```shell
$ make docker-build
$ ./dist/cbnode_<LOCAL_OS>_amd64/cbnode version
```
ez
- Build the binary in the local environment for all supported OS using Goreleaser and check the current version:
```shell
$ make goreleaser
$ ./dist/cbnode_<LOCAL_OS>_amd64/cbnode version
```

### Run

First you must edit the default/template config file into [cmd/cbnode/cfg.buidler.toml](cmd/cbnode/cfg.builder.toml), 
there are more information about the config file into [cmd/cbnode/README.md](cmd/cbnode/README.md)

After setting the config, you can build and run the Chainbing Node as a synchronizer:

```shell
$ make run-node
```

Or build and run as a coordinator, and also passing the config file from other location:

```shell
$ MODE=sync CONFIG=cmd/cbnode/cfg.builder.toml make run-node
```

To check the useful make commands:

```shell
$ make help
```


### Run as a service

```shell
$ sudo make install
```

After, update the config file manually at `/etc/chainbing/config.toml`

```shell
$ sudo service cbnode start
```

To check status

```shell
$ sudo service cbnode status
```

To stop

```shell
$ sudo service cbnode stop
```

If you just kill the process systemd will restart without asking.
