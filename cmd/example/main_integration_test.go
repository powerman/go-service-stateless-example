// +build integration

package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/powerman/check"
	"github.com/powerman/gotest/testexec"
)

func listen() (ln net.Listener, port string) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	_, port, err = net.SplitHostPort(ln.Addr().String())
	if err != nil {
		panic(err)
	}
	return ln, port
}

func TestServeConsulFailed(tt *testing.T) {
	t := check.T(tt)
	ln, port := listen()

	ln.Close()

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	cmd := testexec.Func(ctx, t, main, "-ws.port="+port)
	cmd.Env = append(cmd.Env, "CONSUL_HTTP_ADDR=http://127.0.0.1:"+port)
	out, err := cmd.CombinedOutput()
	t.Match(err, `exit status 1`)
	t.Match(out, `failed to register service`)
}

func TestServeListenFailed(tt *testing.T) {
	t := check.T(tt)
	ln, port := listen()

	defer ln.Close()

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	out, err := testexec.Func(ctx, t, main, "-ws.port="+port).CombinedOutput()
	t.Match(err, `exit status 1`)
	t.Match(out, `address already in use`)
}

func TestServe(tt *testing.T) {
	t := check.T(tt)
	ln, port := listen()

	ln.Close()

	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	out, err := testexec.Func(ctx, t, main, "-ws.port="+port).CombinedOutput()
	t.Match(err, `killed`)
	t.Match(out, `service registered`)
}
