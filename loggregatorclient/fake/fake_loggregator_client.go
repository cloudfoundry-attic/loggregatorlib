package fake

type FakeLoggregatorClient struct {
	Addr     string
	Received chan *[]byte
}

func (flc *FakeLoggregatorClient) Scheme() string {
	return "fake"
}

func (flc *FakeLoggregatorClient) Address() string {
	return flc.Addr
}

func (flc *FakeLoggregatorClient) Send(data []byte) {
	flc.Received <- &data
}

func (flc *FakeLoggregatorClient) Stop() {}
