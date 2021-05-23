# Debugging

This document describes various strategies to debug server problems.

## TLS

To inspect the TLS certificate of the gRPC or JSON-RPC server we can use openssl:

```bash
openssl s_client -servername [your-server-name] -connect 127.0.0.1:4444
```

You can also view the TLS certificate:

```bash
openssl s_client -CAfile settings/dev/certs/root.crt -servername foo.internal-sserver -connect 127.0.0.1:4433 </dev/null 2>/dev/null | openssl x509 -noout -text | grep -B 10 -A 10 DNS:
```

## Servers Using Curl

We can use `curl` to test e.g. the local JSONRPC server. Since the server has a custom root CA we need to pass that to `curl` via the `--cacert` option. We can also tell curl to resolve a given `CNAME` to a local address using the `--resolve` option. For example, to send a request to the JSON-RPC server

```bash
curl --cacert settings/dev/certs/root.crt --resolve jsonrpc-server:5555:127.0.0.1 https://jsonrpc-server:5555
```

This tells CURL to resolve `jsonrpc-server:5555` to `localhost`, which will then make the given `CommonName` of the certificate match what `curl` expects. 

## JSON-RPC Server

To send some example data to the JSON-RPC server via `curl`:

```bash
curl --cacert settings/dev/certs/root.crt --resolve jsonrpc-server:5555:127.0.0.1 https://jsonrpc-server:5555/jsonrpc --header "Content-Type: application/json; charset=utf-8" --data '{"method": "hello", "id": "1", "params": {}, "jsonrpc": "2.0"}' 2>/dev/null | jq .
```