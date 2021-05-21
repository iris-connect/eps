# The EPS TLS-Passthrough Proxy

The EPS server provides a TLS-passthrough proxy service that can forward incoming TLS connections from public clients to an internal server without terminating the TLS connection. The service consists of two servers:

* A **public proxy** that listens on a publicly available TCP port for incoming TLS connections, and on another TCP port for connections from private proxy servers.
* A **private proxy** that does not listen on any TCP port and instead actively connects to the public proxy server when a connection is available for it. It forwards that connection (again without terminating TLS) to an internal server that then handles it.

Both the private and the public proxies have accompanying EPS servers through which they communicate with each other. 

The private proxy can **announce** incoming connections to the public proxy. Whenever a new connection reaches that proxy, it compares it to the announcements it received. If a match is found, it notifies the private proxy of an incoming connection via the EPS network, also sending it a random token. The private proxy opens a connection to the public proxy, sends the token and takes over the TCP stream. It forwards it to an internal server.

## Demo

To demonstrate this mechanism we have prepared an example configuration. Simply run the following snippets in different terminals (from the main directory in the repository):

```bash
# prepare the binaries
make && make examples
# first terminal
internal-server #will open a JSON-RPC server on port 8888
# second terminal (public proxy)
PROXY_SETTINGS=settings/dev/roles/public-proxy-1 proxy run public
# third terminal (private proxy)
PROXY_SETTINGS=settings/dev/roles/private-proxy-1 proxy run private
# fourth terminal (public proxy EPS server)
EPS_SETTINGS=settings/dev/roles/public-proxy-eps-1 eps server run
# fifth terminal (private proxy EPS server)
EPS_SETTINGS=settings/dev/roles/private-proxy-eps-1 eps server run
```

When all services are up and running you should be able to send a request to the proxy via

```bash
curl --cacert settings/dev/certs/root.crt --resolve test.internal-server.local:4433:127.0.0.1 https://test.internal-server.local:4433/jsonrpc | jq .

```

This should return the following JSON data:

```json
{
  "message": "success"
}
```

The request you've sent reached the local TLS server on port 8888 via the two proxies, which communicated through the EPS network to broker the connection. Neat, isn't it?

## Stress Testing

You can also stress-test the server with parallel requests using the `parallel`
util:

```bash
eq 1 2000000 | parallel -j 25 curl --cacert settings/dev/certs/root.crt --resolve test.internal-server.local:4433:127.0.0.1 https://test.internal-server.local:4433/jsonrpc --data "{}"
```

This will try to send 25 requests in parallel to the server.