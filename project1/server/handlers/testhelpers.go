package handlers

import (
	"net"
	"bytes"
	"time"
)

type SendResponseFunc func(conn net.Conn, status string, body string)

type MockConn struct {
	Written bytes.Buffer
}

func (m *MockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *MockConn) Write(b []byte) (n int, err error)  { return m.Written.Write(b) }
func (m *MockConn) Close() error                       { return nil }
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

var (
	testStatus string
	testBody   string
)

func mockSendResponse(conn net.Conn, status string, body string) {
	testStatus = status
	testBody = body
}
