**This software is still a work in progress and not ready for production use!**

# IRIS Endpoint Server (EPS)

This repository contains the code of the IRIS endpoint server (EPS), which manages the communication between different actors in the IRIS ecosystem. It provides a gRPC server & client to exchange messages between different actors, as well as a JSON-RPC API client & server for interacting with the server locally.

## Getting Started

Please ensure your Golang version is recent enough (>=1.13) before you attempt to build the software. 

To build the `eps` binary, simply run

```
make
```

For testing and development you'll also need TLS certificates, which you can generate with

```
make certs
```

Please note that you need `openssl` on your system for this to work. This will generate all required certificates and put them in the `settings/dev/certs` and `settings/dev/test` folders. Please do not use these certificates in a production setting and do not check them into version control.

Please see below for additional dependencies you might need to install for various purposes (e.g. to recompile protobuf code).

To build the example services (e.g. the "locations" services `eps-ls`) simply run

```
make examples
```

## Defining Settings

The `eps` binary will look for settings in a list of colon (`:`) separated directories as defined by the `EPS_SETTINGS` environment variable (or, if it is undefined in the `settings` subdirectory of the current directory). The development settings include an environment-based variable `EPS_OP` that allows you to use different certificates for testing. You should define these variables before running the development server:

```bash
export EPS_SETTINGS=`readlink -f settings/dev`
export EPS_OP=hd-1 # run server as the 'hd-1' operator
```

You can also source these things from the local `.dev-setup` script, which includes everything you need to get started:

```bash
source .dev-setup # load all development environment variables
```

There are also role-specific development/test settings in the `settings/dev/roles` directory. Those can be used to set up multiple EPS servers and test the communication between them. Please have a a look at the [integration guidelines](docs/integration.md) for more information about this.

**Important: The settings parser includes support for variable replacement and many other things. But with great power comes great responsibility and attack surface, so make sure you only feed trusted YAML input to it, as it is not designed to handle untrusted or potentially malicious settings.**

## Running The Server

To run the development EPS server simply run (from the main directory)

```
eps server run
```

For this to work you need to ensure that your `GOPATH` is in your `PATH`. This will open the JSON RPC server and (depending on the settings) also a gRPC server.

## Testing

To run the tests

```
make test # run normal tests
make test-races # test for race conditions
```

## Benchmarks

To run the benchmarks

```
make bench
```

## Debugging

If you're stuck debugging a problem please have a look at the [debugging guidelines](docs/debugging.md), which contain a few pointers that might help you to pinpoint problems in the system.

## Copyright Headers

You can generate and update copyright headers as follows

```
make copyright
```

This will add appropriate headers to all Golang files. You can edit the generation and affected file types directly in the script (in `.scripts`). You should run this before committing code. Please note that any additional comments that appear directly at the top of the file will be replaced by this.

## License

Currently this code is licensed under Affero GPL 3.0.

## Development Requirements

If you make modifications to the protocol buffers (`.proto` files) you need to recompile them using `protoc`. To install this on Debian/Ubuntu systems:

```
sudo apt install protobuf-compiler
```

To generate TLS certificates for testing and development you need to have `openssl` installed.

## Deployment

You can easily deploy the server as a service using `systemd` or Docke. Specific documentation coming up soon.

# Feedback

If you have any questions [just contact us](mailto:iris@steiger-stiftung.de).

# Participation

We are happy about your contribution to the project! In order to ensure compliance with the licensing conditions and the future development of the project, we require a signed contributor license agreement (CLA) for all contributions in accordance with the http://selector.harmonyagreements.org[Harmony standard]. Please sign the corresponding document for .clas/IRIS Gateway-Individual.pdf[natural persons] or for .clas/IRIS Gateway-Entity.pdf[organizations] and send it to [us](mailto:iris@steiger-stiftung.de).

## Supporting organizations

- Björn Steiger Stiftung SbR - https://www.steiger-stiftung.de
