# Peer To Peer (P2P) Proxy

The proxy also supports a peer-to-peer (P2P) mode, which enables two EPS servers to establish a secure connection through the proxy.

## Connection Brokering

The brokering process for a P2P connection is as follows:

* EPS server A wants to connect to EPS server B.
* A recognizes through the directory entry of B that it can only be reached through proxy P.
* A sends a `connectionRequest` message to P over the EPS system, specifying B's gRPC server channel as the recipient.
* P creates a token and send a message to B over the EPS system, forwarding A's request and specifying a proxy endpoint to connect to.
* B receives the message and forwards it to the appropriate channel, which handles it and connects to P's endpoint, sending the token as the routing key.
* P receives the connection from B and stores it.
* P returns a confirmation to A containing the token and same endpoint.
* A connects to P's endpoint and also sends the token.
* P accepts A's connection, retrieves B's matching connection and proxies traffic between them.

## Testing

To set up a test infrastructure, simply run (in different shells):

```bash
# run the service directory
SD_SETTINGS=settings/dev/roles/sd-1 sd run
# run the public proxy
PROXY_SETTINGS=settings/dev/roles/public-proxy-1 proxy run public

# run all EPS servers
EPS_SETTINGS=settings/dev/roles/hd-1 eps server run
EPS_SETTINGS=settings/dev/roles/hd-2 eps server run
EPS_SETTINGS=settings/dev/roles/public-proxy-eps-1 eps server run
```

Make sure you run `make sd-setup` to update the service directory with the necessary entries. Then you should be able to request a ping from the HD-2 server through the proxy via the HD-1 JSON-RPC server:

```bash
curl --cert settings/dev/certs/hd-1.crt --key settings/dev/certs/hd-1.key --cacert settings/dev/certs/root.crt --resolve hd-1:5555:127.0.0.1 https://hd-1:5555/jsonrpc --header "Content-Type: application/json" --data '{"method": "hd-2._ping", "id": "1", "params": {}, "jsonrpc": "2.0"}' | jq .

```