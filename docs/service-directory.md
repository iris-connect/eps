# Service Directory

The service directory is a central database that contains information about all operators in the IRIS ecosystem. It contains information about how operators can be reached and which services they provide.

```yaml
- name: ls-1
  groups: [LocationsServices]
  description: The official "locations service" operated by INöG
  certificates:
    revoked: [a6fcad2145....] # list of revoked certificate IDs
  channels:
    - type: grpc_server
      endpoint: https://ls-1.operators.iris-gateway.de:5555
  services:
    - name: locations
      authorized: [LocationsAdministrators] # global authorization for all group members
      methods:
        - name: get
          authorized: [HealthDepartments]
          async: true
        - name: add
          authorized: [ContactTracingProviders]
          params:
            - name: name
              validators:
                - type: IsString
                  params:
                    min_length: 1
                    max_length: 100
            - name: id
              validators:
                - type: IsBytes
                  params:
                    encoding: base64
                    min_length: 16
                    max_length: 16
- name: recover
  description: The Recover contact tracing app provider
  groups: [ContactTracingProviders, TrustedOperators]
  channels:
    - type: grpc_server
      endpoint: https://recover.operators.iris-gateway.de:5555
  services: [...]
- name: ga-leipzig
  description: The Leipzig health department
  groups: [HealthDepartments]
  channels: [] # no external channels
  services: [] # no external services
```

The directory allows EPS servers to determine whether and how they can connect to another operator. Operators that only have outgoing connectivity (e.g. `ga-leipzig` in the example above) can use the directory to learn that they might receive asynchronous responses from other operators (e.g. `ls-1`) and then open outgoing connections to these operators through which they can receive replies. EPS servers can also use the service directory to determine whether they should accept a message from a given operator.

The service directory implements a group-based permissions mechanism. Currently, only `yes/no` permissions exist (i.e. a member of a given group either can or cannot call a given service method). More fine-grained permissions (e.g. a contact tracing provider can only edit its own entries in the "locations" service) need to be implemented by the services themselves. For that purpose, the EPS server makes information about the calling peer available to the services via a special parameter (`_caller`) that gets passed along with the other RPC method parameters. This structure also contains the current entry of the caller from the service directory, making it easy for the called service to identify and authorize the caller.

## Service Directory API

The EPS server package also provides a `sd` API server command that opens a JSON-RPC server which distributes the service directory.

```bash

```

## Signature Scheme

All entries in the service directory should be cryptographically signed. For this, every actor in the EPS system has a pair of ECDSA keys and an accompanying certificate.