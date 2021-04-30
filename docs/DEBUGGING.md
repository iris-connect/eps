# Debugging

This document describes various strategies to debug server problems.

## TLS

To inspect the TLS certificate of the gRPC or JSON-RPC server we can use openssl:

```
openssl s_client -connect 127.0.0.1:4444
```