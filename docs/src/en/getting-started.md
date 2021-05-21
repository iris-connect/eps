# Getting Started

Integrating with the IRIS infrastructure using the EPS server is easy (we hope). First, you need to deploy the `eps` server together with the settings and certificates we've provided to you. This is as easy as downloading the latest `eps` version from our server, unpacking the settings archive we've provided you with and running

```bash
EPS_SETTINGS=path/to/settings eps server run
```

This should open a local JSON-RPC server on port `5555` that you can connect to via TLS (you'll need to add the root CA certificate to your certificate chain for this). This server is your gateway to all IRIS services. Simply look up the services that a specific operator provides and send a request that contains the name of the operator and the service method you want to call. For example, to interact with a "locations" service provided by operator "ls-1", you'd simply post a JSON RPC message like this:

```json
{
	"method": "ls-1.add",
	"id": "1",
	"params": {
		"name": "Ginos",
		"id": "af5ca4da5caa"
	},
	"jsonrpc": "2.0"
}
```

The gateway will take care of routing this message to the correct service and returning a response to you.

If you want to accept requests from other services in the IRIS ecosystem you can use the `jsonrpc_client`, simply specifying an API endpoint that incoming requests will be delivered to using the same syntax as above.

That's it!

## Asynchronous Calls

The calls we've seen above were all synchronous, i.e. making a call resulted in a direct response. Sometimes calls need to be asynchronous though, e.g. because replying to them takes time. If you make an asynchronous call to another service, you'll get back an acknowledgment first. As soon as the service you've called has a response ready, it will send it back to your via the `eps` network, using the same `id` you provided (which enables you to match the response to your request). Likewise, you can respond to calls from other services in an asynchronous way, simply pushing the response to your local JSON-RPC server with a method name `respond` (without a service name). Do not forget to include the same `id` that you received with the original request, as this will contain the "return address" of the request.

## Integration Example

To get a concrete idea of how to integrate with the IRIS infrastructure using the EPS server we have created a simple demo setup that illustrates all components. The demo consists of three components:

* An `eps` server simulating a `health department`, named `hd-1`
* An `eps` server simulation an operator offering a "locations" service, named `ls-1`
* The actual location service `eps-ls` offered by the operator `ls-1`

## Getting Up And Running

First, please check the README on how to create all necessary TLS certificates and build the software. Then, start the individual services on different terminals:

```bash
# run the `eps` server of the "locations" operator ls-1
EPS_SETTINGS=settings/dev/roles/ls-1 eps --level debug server run
# run the `eps` server of the health department hd-1
EPS_SETTINGS=settings/dev/roles/hd-1 eps --level debug server run
# run the "locations" service
eps-ls
```

## Testing

Now your system should be up and running. The demo "locations" service provides a simple, authenticationless JSON-RPC interface with two methods: `add`, which will add a location to the database, and `lookup`, which will look up a location based on its `name`. For example, to add a location to the service:

```bash
curl --cacert settings/dev/certs/root.crt --resolve hd-1:5555:127.0.0.1 https://hd-1:5555/jsonrpc --header "Content-Type: application/json" --data '{"method": "ls-1.add", "id": "1", "params": {"name": "Ginos", "id": "af5ca4da5caa"}, "jsonrpc": "2.0"}' 2>/dev/null | jq 
```

This should return a simple JSON response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "_": "ok"
  },
  "id": "1"
}
```

The request first went to the health department's `eps` server, was first routed to `ls-1`'s `eps` server via gRPC and was then passed to the JSON-RPC API of the local `eps-ls` service running there. The result was then passsed back along the entire chain.

You can also perform a lookup of the location you've just added:

```bash
curl --cacert settings/dev/certs/root.crt --resolve hd-1:5555:127.0.0.1 https://hd-1:5555/jsonrpc --header "Content-Type: application/json" --data '{"method": "ls-1.lookup", "id": "1", "params": {"name": "Ginos"}, "jsonrpc": "2.0"}' 2>/dev/null | jq .
```

which should return

```json
{
  "jsonrpc": "2.0",
  "result": {
    "id": "af5ca4da5caa",
    "name": "Ginos"
  },
  "id": "1"
}
```

Hence, interacting with the remote "locations" service is just like calling a local JSON-RPC service, except that you specify the name of the operator running the service, `ls-1.lookup`, instead of just calling `lookup`.
