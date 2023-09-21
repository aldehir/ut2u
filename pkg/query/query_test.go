package query

import (
	"context"
	"net"
	"testing"
)

func TestClientQuery(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Error(err)
	}
	defer client.Close()

	addr, err := net.ResolveUDPAddr("udp", "nj-1.kokuei.dev:7778")
	if err != nil {
		t.Error(err)
	}

	details, err := client.Query(context.TODO(), addr, WithRules(), WithPlayers())
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", details)
}
