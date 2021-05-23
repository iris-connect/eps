# Public Proxy

The public proxy receives TLS connections from the public Internet, analyzes
the Server Name Indication (SNI) header and forwards the TLS connection to
the internal proxy, which terminates TLS and forwards the connection to
an internal service that processes it.

The public proxy provides a JSON-RPC server that the accompanying EPS server connects to. It also has a JSON-RPC client to connect to that EPS server.