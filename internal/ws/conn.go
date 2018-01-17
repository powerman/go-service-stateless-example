package ws

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/powerman/must"
	"github.com/powerman/structlog"
)

var errSlowClient = errors.New("client is reading too slow")

type clientConn struct {
	context.Context
	cancel context.CancelFunc
	conn   *websocket.Conn
	send   chan *websocket.PreparedMessage
	log    *structlog.Logger
}

// Log return logger configured to output connection details.
func (c *clientConn) Log() *structlog.Logger {
	return c.log
}

// Write will try to send WHOLE p to connection without blocking.
// If this is not possible then connection will be closed.
//
// Write will copy p, so if you need to send same message to many
// connections then use WritePrepared instead.
//
// Safe for simultaneous use by multiple goroutines.
func (c *clientConn) Write(p []byte) (n int, err error) {
	pm, err := websocket.NewPreparedMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, c.log.Err(err)
	}

	err = c.WritePrepared(pm)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// WritePrepared will try to send pm to connection without blocking.
// If this is not possible then connection will be closed.
//
// Safe for simultaneous use by multiple goroutines.
func (c *clientConn) WritePrepared(pm *websocket.PreparedMessage) error {
	select {
	case <-c.Done():
		c.log.Debug(c.Err())
		return c.Err()
	case c.send <- pm:
		return nil
	default:
		_ = c.Close()
		return c.log.Err(errSlowClient)
	}
}

// Close connection and cancel it context. Can be called multiple times,
// but only first call may return some error.
//
// Safe for simultaneous use by multiple goroutines.
func (c *clientConn) Close() error {
	c.cancel()

	err := c.conn.Close()
	if err != nil {
		// Without adding extra mutex we can't guarantee conn.Close() will
		// be called just once, so repeated calls is not an error.
		if isClosed(err) {
			return nil
		}
		return c.log.Err("failed to close", "err", err)
	}

	c.log.Info("closed")
	return nil
}

func (c *clientConn) writer() {
	defer c.Close()

	var err error
	for err == nil {
		select {
		case <-c.Done():
			return
		case pm := <-c.send:
			must.NoErr(c.conn.SetWriteDeadline(time.Now().Add(writeTimeout)))
			err = c.conn.WritePreparedMessage(pm)
		}
	}
	c.logErr("write", err)
}

func (c *clientConn) reader() {
	log := c.log

	defer c.Close()

	c.conn.SetReadLimit(cfg.readLimit)
	must.NoErr(c.conn.SetReadDeadline(time.Now().Add(readTimeout)))
	c.conn.SetPingHandler(func(string) error {
		must.NoErr(c.conn.SetReadDeadline(time.Now().Add(readTimeout)))
		must.NoErr(c.conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(writeTimeout)))
		return nil
	})

	var err error
	for err == nil {
		var msg []byte
		_, msg, err = c.conn.ReadMessage()
		if err != nil {
			c.logErr("read", err)
		} else {
			log.Info("recv", "msg", string(msg))
			msg = append([]byte("echo: "), msg...)
			_, err = c.Write(msg)
		}
	}
}

func (c *clientConn) logErr(op string, err error) {
	log := c.log

	if e, ok := err.(*websocket.CloseError); ok {
		if e.Code == websocket.CloseGoingAway {
			log.Debug("client go away")
		} else {
			log.Err("client closed", "err", err)
		}
	} else if isTimeout(err) {
		log.Err(op + " timeout")
	} else if !isClosed(err) {
		log.Err("failed to "+op, "err", err)
	}
}

// https://github.com/golang/go/issues/4373#issuecomment-352964424
func isClosed(err error) bool {
	const errNetClosing = "use of closed network connection" // from internal/poll/fd.go
	if e, ok := err.(*net.OpError); ok {
		return e.Err.Error() == errNetClosing
	}
	return false
}

func isTimeout(err error) bool {
	if e, ok := err.(net.Error); ok {
		return e.Timeout()
	}
	return false
}
