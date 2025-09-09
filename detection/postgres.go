package detection

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
)

var pqDialer net.Dialer

func IsPostgresql(ctx context.Context, host string, port int) (bool, error) {
	conn, err := pqDialer.DialContext(ctx, "tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return false, err
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
	c, err := conn.Write(inMessage)
	if err != nil {
		return false, err
	}
	if c != len(inMessage) {
		return false, fmt.Errorf("written %d of %d", c, len(inMessage))
	}
	var buffer [11]byte

	toRecv := []byte{
		0x45, 0x00, 0x00, 0x00, 0x85, 0x53, 0x46, 0x41, 0x54, 0x41, 0x4c,
	}

	n, err := io.ReadAtLeast(conn, buffer[:], len(toRecv))
	if err != nil {
		return false, err
	}
	if n != len(toRecv) {
		return false, fmt.Errorf("read %d of %d", n, len(toRecv))
	}
	if !bytes.Equal(buffer[:n], toRecv) {
		return false, nil
	}
	return true, nil
}
