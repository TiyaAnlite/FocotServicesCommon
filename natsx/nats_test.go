package natsx

import "testing"

func TestNATS(t *testing.T) {
	nats := NatsHelper{}
	if err := nats.Open(NatsConfig{}); err != nil {
		t.Fatal(err)
	}
}
