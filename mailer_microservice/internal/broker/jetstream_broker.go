package broker

import (
	"github.com/nats-io/nats.go"
)

type JetStreamClient struct {
	js nats.JetStreamContext
}

func (c *JetStreamClient) GetJetStream() nats.JetStreamContext {
	return c.js
}

func NewJetStreamClient(conn *nats.Conn) (*JetStreamClient, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}
	return &JetStreamClient{js: js}, nil
}

func (c *JetStreamClient) Publish(subject string, data []byte) error {
	_, err := c.js.Publish(subject, data)
	return err
}

// Subscribe sets up a durable JetStream consumer with manual ack.
// handler is responsible for calling msg.Ack() or msg.Nak() explicitly.
func (c *JetStreamClient) Subscribe(subject, durable string, handler func(msg *nats.Msg)) error {
	_, err := c.js.Subscribe(subject, handler,
		nats.Durable(durable),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.MaxDeliver(5),
	)
	return err
}

func (c *JetStreamClient) EnsureStream(streamName string, subjects []string) error {
	_, err := c.js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: subjects,
		Storage:  nats.FileStorage,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return err
	}
	return nil
}
