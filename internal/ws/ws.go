// Package ws implements WebSocket API.
package ws

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/powerman/structlog"

	"github.com/powerman/go-service-stateless-example/internal/logkey"
)

// Constants defined as vars for tests.
var (
	readTimeout = 30 * time.Second
)

const (
	writeTimeout = 5 * time.Second
	// To minimize amount of read/write syscalls bufsize should
	// provide enough space for at least one average-size message.
	// WebSocket overhead is up to 14 bytes.
	// So 512-byte buffer may contain about 120 4-byte utf-8 symbols.
	bufsize      = 512
	sendChanSize = 8
)

var (
	log = structlog.New()
	cfg struct {
		readLimit int64
	}
	upgrader = websocket.Upgrader{
		// CheckOrigin: func(r *http.Request) bool { return true },
		HandshakeTimeout: writeTimeout,
		// Set buffer size to replace 4096-byte net/http buffers.
		ReadBufferSize:  bufsize,
		WriteBufferSize: bufsize - 14, // subtract maxFrameHeaderSize:
		// https://github.com/gorilla/websocket/blob/master/conn.go#L31
	}
)

// Init must be called once before using this package.
//
// We may have to read about readLimit bytes before we'll notice attempt
// to send more (in case of WebSocket fragmented message), so try to keep
// readLimit small enough (about several KB).
func Init(readLimit int64) error {
	cfg.readLimit = readLimit
	return nil
}

// Serve accepts WebSocket connections.
func Serve(w http.ResponseWriter, r *http.Request) {
	remote := fmt.Sprintf("%15s:%-5s", r.Header.Get("X-Real-IP"), r.Header.Get("X-Real-Port"))
	log := log.New(logkey.Remote, remote)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Err("failed to upgrade connection", "err", err)
		return
	}
	log.Info("connected")

	ctx, cancel := context.WithCancel(context.Background())
	c := &clientConn{
		Context: ctx,
		cancel:  cancel,
		conn:    conn,
		send:    make(chan *websocket.PreparedMessage, sendChanSize),
		log:     log,
	}
	go c.writer()
	go c.reader()
}
