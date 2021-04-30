package grpc

import (
	"context"
	cryptoTls "crypto/tls"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/protobuf"
	"github.com/iris-gateway/eps/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

type Client struct {
	tlsConfig  *cryptoTls.Config
	address    string
	serverName string
	connection *grpc.ClientConn
	settings   *eps.GRPCClientSettings
}

func (c *Client) Connect() error {
	var err error

	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConfig)))

	c.connection, err = grpc.Dial(c.address, opts...)

	if err != nil {
		return err
	}

	return nil

}

func (c *Client) Close() error {
	return c.connection.Close()
}

func (c *Client) SendMessage() error {

	client := protobuf.NewEPSClient(c.connection)

	message := &protobuf.Message{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	msgClient, err := client.MessageExchange(ctx)

	if err != nil {
		return err
	}

	msgClient.Send(message)

	return nil

}

func MakeClient(settings *eps.GRPCClientSettings, address, serverName string) (*Client, error) {

	tlsConfig, err := tls.TLSClientConfig(settings.TLS, serverName)

	if err != nil {
		return nil, err
	}

	return &Client{
		tlsConfig:  tlsConfig,
		settings:   settings,
		address:    address,
		serverName: serverName,
	}, nil

}
