package provider

import (
	"bufio"
	"context"
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
