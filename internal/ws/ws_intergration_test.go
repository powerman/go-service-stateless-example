// +build integration

package ws

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/powerman/check"
)

func TestRemote(tt *testing.T) {
	t := check.T(tt)

	conn, _, err := dialer.Dial("ws://example:8080", nil)
	t.Nil(err)

	t.Nil(conn.SetWriteDeadline(deadline()))
	t.Nil(conn.WriteMessage(websocket.TextMessage, []byte(`hello`)))
	t.Nil(conn.SetReadDeadline(deadline()))
	msgType, p, err := conn.ReadMessage()
	t.Nil(err)
	t.Equal(msgType, websocket.TextMessage)
	t.Equal(string(p), `echo: hello`)
	t.Nil(conn.Close())
}
