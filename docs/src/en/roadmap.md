# Roadmap

* Full service directory functionality
* Rate-limiting & advanced security features
* Data signing for the service directory
* Extensive unit tests for all components

## gRPC Client Channel

* Listen for changes in the service directory and update outgoing connections accordingly.

## gRPC Server Channel

## JSON-RPC Client Channel

* Add a way to specify which services are reachable via the channel to e.g. enable an operator of an EPS server to define multiple JSON-RPC clients for different services in the infrastructure.

## Service Directory API

* Persistence layer to store signed changes.
* A way to verify signatures.s
