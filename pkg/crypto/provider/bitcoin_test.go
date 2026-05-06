package provider

import (
	"bufio"
	context "context"
	"net"
	"runtime"
	"testing"
	"time"
)

func TestElectrumClientRequestWithContextDoesNotLeakGoroutines(t *testing.T) {
	url, stopServer := startElectrumTestServer(t)
	defer stopServer()

	client, err := NewElectrumClient(url)
	if err != nil {
		t.Fatalf("expected electrum client creation to succeed, got: %v", err)
	}

	ctx := context.Background()
	baseline := runtime.NumGoroutine()

	for i := 0; i < 30; i++ {
		_, err := client.RequestWithContext(ctx, Request{
			ID:     int64(i + 1),
			Method: "server.ping",
			Params: []interface{}{},
		})
		if err != nil {
			t.Fatalf("expected request %d to succeed, got: %v", i+1, err)
		}
	}

	maxExpectedGoroutines := baseline + 8
	deadline := time.Now().Add(2 * time.Second)
	for {
		runtime.GC()
		current := runtime.NumGoroutine()
		if current <= maxExpectedGoroutines {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf(
				"expected request goroutines to exit; baseline=%d current=%d",
				baseline,
				current,
			)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func startElectrumTestServer(t *testing.T) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("expected test server listener to start, got: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					t.Logf("test electrum server accept failed: %v", err)
					return
				}
			}

			go func(conn net.Conn) {
				defer conn.Close()

				if _, err := bufio.NewReader(conn).ReadBytes('\n'); err != nil {
					return
				}
				_, _ = conn.Write([]byte(`{"id":1,"result":true,"error":null}` + "\n"))
			}(conn)
		}
	}()

	return "tcp://" + listener.Addr().String(), func() {
		close(done)
		_ = listener.Close()
	}
}
