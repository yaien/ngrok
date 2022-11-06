package ngrok

import (
	"context"
	"testing"
)

func TestOpen(t *testing.T) {
	tunnel, err := Open(context.TODO(), Options{Addr: ":3000"})
	if err != nil {
		t.Fatalf("failed oppenning tunnel: %s", err)
	}
	defer tunnel.Close()
	t.Log("Tunnel listening on", tunnel.Url())
}
