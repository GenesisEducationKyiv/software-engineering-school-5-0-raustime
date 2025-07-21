package broker

import (
	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	conn *nats.Conn
}

type Message struct {
	Subject string
	Data    []byte
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

func (n *NATSClient) Subscribe(subject string, handler func(*Message)) (*nats.Subscription, error) {
	return n.conn.Subscribe(subject, func(m *nats.Msg) {
		handler(&Message{Subject: m.Subject, Data: m.Data})
	})
}

func (n *NATSClient) Close() {
	n.conn.Close()
}
