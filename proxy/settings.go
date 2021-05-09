package proxy

import (
	"github.com/iris-gateway/eps/jsonrpc"
)

type Settings struct {
	Private *PrivateServerSettings `json:"private"`
	Public  *PublicServerSettings  `json:"public"`
}

type PublicServerSettings struct {
	BindAddress   string                         `json:"bind_address"`
	EPSEndpoint   string                         `json:"eps_endpoint"`
	JSONRPCServer *jsonrpc.JSONRPCServerSettings `json:"jsonrpc_server`
}

type PrivateServerSettings struct {
	BindAddress   string                         `json:"bind_address"`
	EPSEndpoint   string                         `json:"eps_endpoint"`
	JSONRPCServer *jsonrpc.JSONRPCServerSettings `json:"jsonrpc_server`
}
