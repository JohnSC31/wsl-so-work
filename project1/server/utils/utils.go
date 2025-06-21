package utils

import (
	"bytes"
	"net"
	"time"
)

type FakeConn struct {
	Buffer *bytes.Buffer
}

func (f *FakeConn) Write(p []byte) (n int, err error) {
	return f.Buffer.Write(p)
}

func (f *FakeConn) Close() error {
	return nil
}

// Implementa otros métodos necesarios de net.Conn
func (f *FakeConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (f *FakeConn) LocalAddr() net.Addr                { return nil }
func (f *FakeConn) RemoteAddr() net.Addr               { return nil }
func (f *FakeConn) SetDeadline(t time.Time) error      { return nil }
func (f *FakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (f *FakeConn) SetWriteDeadline(t time.Time) error { return nil }
