# Debugging

This document describes various strategies to debug server problems.

## TLS

To inspect the TLS certificate of the gRPC or JSON-RPC server we can use openssl:

```
openssl s_client -connect 127.0.0.1:4444
```

## Servers Using Curl

We can use `curl` to test e.g. the local JSONRPC server. Since the server has a custom root CA we need to pass that to `curl` via the `--cacert` option. We can also tell curl to resolve a given `CNAME` to a local address using the `--resolve` option. For example, to send a request to the JSON-RPC server

```bash
curl --cacert settings/dev/certs/rootCA.crt --resolve jsonrpc-server:5555:127.0.0.1 https://jsonrpc-server:5555
```

This tells CURL to resolve `jsonrpc-server:5555` to `localhost`, which will then make the given `CommonName` of the certificate match what `curl` expects. 