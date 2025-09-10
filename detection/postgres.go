package detection

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
)

var pqDialer net.Dialer

// IsPostgresql detects if there is a postgres instance on specified host/port
//
// It roughly consumes 1Kb of memory per call
func IsPostgresql(ctx context.Context, host string, port int) (bool, error) {
	conn, err := pqDialer.DialContext(ctx, "tcp4", fmt.Sprintf("%s:%v", host, port))
	if err != nil {
		return false, nil
	}
	defer conn.Close()
	inMessage := []byte{
		0x00,
		0x00,
		0x00,
		0x08,
		0x00,
		0x03,
		0x00,
		0x00,
	}

	inMessage = append(inMessage, []byte("user\x00")...)
	inMessage = append(inMessage, []byte("user\x00")...)
	inMessage = append(inMessage, []byte("database\x00")...)
	inMessage = append(inMessage, []byte("database\x00\x00")...)

	messageLen := len(inMessage)
	inMessage[3] = byte(messageLen)

	c, err := conn.Write(inMessage[:])
	if err != nil {
		return false, nil
	}
	if c != len(inMessage) {
		return false, fmt.Errorf("written %d of %d", c, len(inMessage))
	}
	var buffer [4]byte

	toRecv := [...]byte{
		0x52, 0x00, 0x00, 0x00,
	}

	var n int

	done := make(chan bool)

	go func() {
		defer close(done)
		n, err = io.ReadAtLeast(conn, buffer[:], len(toRecv))
		done <- true
	}()

	select {
	case <-ctx.Done():
		return false, nil
	case <-done:

	}
	if err != nil {
		return false, err
	}
	if n != len(toRecv) {
		return false, nil
	}
	if !bytes.Equal(buffer[:n], toRecv[:]) {
		return false, nil
	}
	return true, nil
}
