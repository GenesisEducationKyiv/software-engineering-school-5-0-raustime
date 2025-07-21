package broker

type MockBroker struct {
	LastSubject string
	LastData    []byte
}

func (m *MockBroker) Publish(subject string, data []byte) error {
	m.LastSubject = subject
	m.LastData = data
	return nil
}
