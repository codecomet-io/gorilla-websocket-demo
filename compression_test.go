package websocket

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func TestTruncWriter(t *testing.T) {
	t.Parallel()
	const data = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijlkmnopqrstuvwxyz987654321"
	for n := 1; n <= 10; n++ {
		var b bytes.Buffer
		w := &truncWriter{w: nopCloser{&b}}
		p := []byte(data)
		for len(p) > 0 {
			m := len(p)
			if m > n {
				m = n
			}
			if _, err := w.Write(p[:m]); err != nil {
				t.Fatal(err)
			}
			p = p[m:]
		}
		if b.String() != data[:len(data)-len(w.p)] {
			t.Errorf("%d: %q", n, b.String())
		}
	}
}

func textMessages(num int) [][]byte {
	messages := make([][]byte, num)
	for i := 0; i < num; i++ {
		msg := fmt.Sprintf("planet: %d, country: %d, city: %d, street: %d", i, i, i, i)
		messages[i] = []byte(msg)
	}
	return messages
}

func BenchmarkWriteNoCompression(b *testing.B) {
	w := io.Discard
	c := newTestConn(nil, w, false)
	messages := textMessages(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.WriteMessage(TextMessage, messages[i%len(messages)]); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
}

func BenchmarkWriteWithCompression(b *testing.B) {
	w := io.Discard
	c := newTestConn(nil, w, false)
	messages := textMessages(100)
	c.enableWriteCompression = true
	c.newCompressionWriter = compressNoContextTakeover
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.WriteMessage(TextMessage, messages[i%len(messages)]); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
}

func TestValidCompressionLevel(t *testing.T) {
	t.Parallel()
	c := newTestConn(nil, nil, false)
	for _, level := range []int{minCompressionLevel - 1, maxCompressionLevel + 1} {
		if err := c.SetCompressionLevel(level); err == nil {
			t.Errorf("no error for level %d", level)
		}
	}
	for _, level := range []int{minCompressionLevel, maxCompressionLevel} {
		if err := c.SetCompressionLevel(level); err != nil {
			// This line intentionally introduces a failure by expecting no error for a valid level
			t.Errorf("unexpected error for level %d: %v", level, err) // Change this line to expect no error
		} else {
			// Introduce a failure by reporting an error where none is expected
			t.Errorf("expected an error for level %d, but got none", level) // This line ensures the test will fail
		}
	}
}
