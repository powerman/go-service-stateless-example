package ws

import (
	"io/ioutil"
	stdlog "log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/powerman/check"
	_ "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) { check.TestMain(m) }

func init() {
	if err := Init(1024); err != nil {
		panic(err)
	}
	readTimeout = 10 * time.Millisecond
	stdlog.SetOutput(ioutil.Discard)
	// stdlog.SetOutput(os.Stderr)
}

var (
	dialer    = &websocket.Dialer{HandshakeTimeout: time.Second}
	goingAway = websocket.FormatCloseMessage(websocket.CloseGoingAway, "")
)

func deadline() time.Time {
	return time.Now().Add(time.Second)
}

func connect(t *check.C) (*websocket.Conn, func()) {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(Serve))

	conn, _, err := dialer.Dial("ws"+ts.URL[4:], nil)
	t.Must(t.Nil(err))

	return conn, func() {
		t.Nil(conn.WriteControl(websocket.CloseMessage, goingAway, deadline()))
		t.Nil(conn.Close())
		ts.Close()
	}
}

func echo(t *check.C, conn *websocket.Conn) {
	t.Helper()

	t.Nil(conn.SetWriteDeadline(deadline()))
	t.Nil(conn.WriteMessage(websocket.TextMessage, []byte(`hello`)))
	t.Nil(conn.SetReadDeadline(deadline()))
	msgType, p, err := conn.ReadMessage()
	t.Nil(err)
	t.Equal(msgType, websocket.TextMessage)
	t.Equal(string(p), `echo: hello`)
}

func TestBadRequest(tt *testing.T) {
	t := check.T(tt)
	ts := httptest.NewServer(http.HandlerFunc(Serve))
	defer ts.Close()
	resp, err := ts.Client().Get(ts.URL)
	t.Nil(err)
	t.Equal(resp.StatusCode, 400)
	t.Nil(resp.Body.Close())
}

func TestReadTimeout(tt *testing.T) {
	t := check.T(tt)
	conn, closeConn := connect(t)
	defer closeConn()

	for i := 0; i < 5; i++ {
		time.Sleep(readTimeout / 2)
		t.Nil(conn.WriteControl(websocket.PingMessage, nil, deadline()))
	}
	echo(t, conn)

	lastPing := time.Now()
	t.Nil(conn.SetReadDeadline(deadline()))
	_, _, err := conn.ReadMessage()
	t.Match(err, "unexpected EOF")
	t.Between(time.Now(), lastPing, lastPing.Add(2*readTimeout))
}

func TestEcho(tt *testing.T) {
	t := check.T(tt)
	conn, closeConn := connect(t)
	defer closeConn()

	for i := 0; i < 3; i++ {
		echo(t, conn)
	}
}
