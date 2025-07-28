package broker

import (
	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	conn *nats.Conn
}

func NewNATSClient(url string) (*NATSClient, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NATSClient{conn: conn}, nil
}

func (n *NATSClient) Publish(subject string, data []byte) error {
	return n.conn.Publish(subject, data)
}

func (n *NATSClient) Close() {
	n.conn.Close()
}
