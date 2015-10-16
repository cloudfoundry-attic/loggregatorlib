package fake

type FakeLoggregatorClient struct {
	Received chan *[]byte
}

func (flc FakeLoggregatorClient) Send(data []byte) {
	flc.Received <- &data
}

func (flc FakeLoggregatorClient) Stop() {}
