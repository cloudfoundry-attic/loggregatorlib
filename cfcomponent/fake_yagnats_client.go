package cfcomponent

import "github.com/cloudfoundry/yagnats"

type FakeYagnatsClient struct {
}

func (c *FakeYagnatsClient) Ping() bool {
	return true
}

func (c *FakeYagnatsClient) Connect(connectionProvider yagnats.ConnectionProvider) error {
	return nil
}

func (c *FakeYagnatsClient) Disconnect() {
}

func (c *FakeYagnatsClient) Publish(subject string, payload []byte) error {
	return nil
}

func (c *FakeYagnatsClient) PublishWithReplyTo(subject, reply string, payload []byte) error {
	return nil
}

func (c *FakeYagnatsClient) Subscribe(subject string, callback yagnats.Callback) (int, error) {
	return 0, nil
}

func (c *FakeYagnatsClient) SubscribeWithQueue(subject, queue string, callback yagnats.Callback) (int, error) {
	return 0, nil
}

func (c *FakeYagnatsClient) Unsubscribe(subscription int) error {
	return nil
}

func (c *FakeYagnatsClient) UnsubscribeAll(subject string) {
}
