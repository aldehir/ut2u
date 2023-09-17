package query

import (
	"context"
	"testing"
	"time"
)

func TestQueryPing(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := QueryPing(ctx, "chi-1.staging.kokuei.dev:7778")
	if err != nil {
		t.Error(err)
	}

	t.Logf("Resp: %+v\n", resp)
}

func TestQueryRules(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := QueryRules(ctx, "chi-1.staging.kokuei.dev:7778")
	if err != nil {
		t.Error(err)
	}

	t.Logf("Resp: %+v\n", resp)
}
