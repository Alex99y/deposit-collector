package provider

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"testing"
)

func TestBitcoinProviderBroadcastTransactionReturnsUnquotedTxHash(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected listener to start, got: %v", err)
	}
	defer listener.Close()

	const txHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		var req Request
		if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil {
			serverErr <- err
			return
		}
		if req.Method != "blockchain.transaction.broadcast" {
			t.Errorf("expected broadcast method, got %q", req.Method)
		}

		_, err = conn.Write([]byte(`{"id":1,"result":"` + txHash + `","error":null}` + "\n"))
		serverErr <- err
	}()

	provider := BitcoinProvider{
		electrumClient: ElectrumClient{url: "tcp://" + listener.Addr().String()},
		context:        context.Background(),
	}

	actualTxHash, err := provider.BroadcastTransaction("signed-tx-hex")
	if err != nil {
		t.Fatalf("expected broadcast to succeed, got: %v", err)
	}
	if actualTxHash != txHash {
		t.Fatalf("expected tx hash %q, got %q", txHash, actualTxHash)
	}

	if err := <-serverErr; err != nil {
		t.Fatalf("fake electrum server failed: %v", err)
	}
}
