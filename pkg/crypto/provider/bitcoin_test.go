package provider

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestElectrumClientRequestWithContextDoesNotLeakGoroutines(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				if _, err := reader.ReadBytes('\n'); err != nil {
					return
				}
				_, _ = fmt.Fprintln(conn, `{"id":1,"result":42,"error":null}`)
			}(conn)
		}
	}()

	client := &ElectrumClient{url: "tcp://" + listener.Addr().String()}
	before := runtime.NumGoroutine()

	for i := 0; i < 25; i++ {
		_, err := client.RequestWithContext(ctx, Request{
			ID:     int64(i),
			Method: "server.ping",
			Params: []interface{}{},
		})
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}

	var after int
	for i := 0; i < 20; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
		after = runtime.NumGoroutine()
		if after <= before+5 {
			return
		}
	}

	t.Fatalf(
		"goroutines grew after completed requests: before=%d after=%d stacks:\n%s",
		before,
		after,
		goroutineStacks(),
	)
}

func goroutineStacks() string {
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	stacks := string(buf[:n])
	lines := strings.Split(stacks, "\n")
	if len(lines) > 200 {
		lines = lines[:200]
	}
	return strings.Join(lines, "\n")
}

func TestBitcoinProviderBroadcastTransactionReturnsUnquotedTxHash(
	t *testing.T,
) {
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

		_, err = conn.Write(
			[]byte(`{"id":1,"result":"` + txHash + `","error":null}` + "\n"),
		)
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
