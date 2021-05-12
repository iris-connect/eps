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

## Signature Schema

All entries in the service directory should be cryptographically signed. For this, every actor in the EPS system has a pair of ECDSA keys and an accompanying certificate.

### Signing Data

The `sdh` tool includes a `sign` command that allows us to sign arbitrary JSON data. It uses the signing signatures generated by the `make certs` Make command. For example, to sign a JSON file, simply use

```
# define the SD settings
export SD_SETTINGS=settings/dev/roles/private-proxy-1/sdh

#sign a service directory entry
sdh sign settings/dev/roles/private-proxy-1/sdh/entry.json
```

The output should e.g. look like this:

```json
{
  "signature": {
    "r": "67488385997031737348502334621054744305438368369525250023542608571625588981387",
    "s": "110557266089828975725234959115295121652814407881082688883738138814924173982570",
    "c": "-----BEGIN CERTIFICATE-----\nMIIC1TCCAb2gAwIBAgIUe3+081Bi4Z0DXDdeBhfZZOAs4OwwDQYJKoZIhvcNAQEL\nBQAwaTELMAkGA1UEBhMCREUxDzANBgNVBAgMBkJlcmxpbjEPMA0GA1UEBwwGQmVy\nbGluMQ0wCwYDVQQKDARJUklTMQswCQYDVQQLDAJJVDEcMBoGA1UEAwwTVGVzdGlu\nZy1EZXZlbG9wbWVudDAeFw0yMTA1MTExMTMzNDBaFw0yMjA5MjMxMTMzNDBaMGUx\nCzAJBgNVBAYTAkRFMQ8wDQYDVQQIDAZCZXJsaW4xDzANBgNVBAcMBkJlcmxpbjEN\nMAsGA1UECgwESVJJUzELMAkGA1UECwwCSVQxGDAWBgNVBAMMD3ByaXZhdGUtcHJv\neHktMTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABLHlILI5POvEDJc96W0dbag7\nFt8BVmitGqwS5jarYRwOUe/PiQ8tMBkMw9X/2U8G1qGYQb/CiRDh1DDy/Eh/mGKj\nRDBCMDMGA1UdEQQsMCqCD3ByaXZhdGUtcHJveHktMYIXKi5wcml2YXRlLXByb3h5\nLTEubG9jYWwwCwYDVR0PBAQDAgeAMA0GCSqGSIb3DQEBCwUAA4IBAQAmUESzD1ls\nmpECtRlinhiUduif9nVddtLeW/Ui86PHkS50vjSOVHY7ZHrfWbFB4/p4bwm8Sp1/\npFHx4WyuHiow5Ah3HV9afDcgyWBd1V8ijIFOlNF27u/caVsa9gV7iDVJ+6mBXKkf\nCgNI2bA2WoOVXQMwRoow4vSYrVAdM/Eyq8PHYOHkGqdd4uASG5df4vE+gnB2z9WD\nFuxkVYkncVP5OB+N7EAkQrVjrITdiSN0yYAVWFKz1IEnPF7GRW6KsPHW9lJeePeD\n1gLNh2KF6drrXT2PIIYVB31uepSoCqFnUUDcC/PX0qHu8jilvr/pTzhFUWbuX+Ja\nfaIRxqWB0frZ\n-----END CERTIFICATE-----\n"
  },
  "data": {
    "foo": "bar",
    "name": "private-proxy-1"
  }
}
```

Before importing such data, we can check the signature using the `verify` command (you need to specify the expected `name` of the signer):

```
export SD_SETTINGS=settings/dev/roles/sd-1/sdh
sdh verify signed.json private-proxy-1
```

If the signature is valid the exit code will be `0`, otherwise `1`.