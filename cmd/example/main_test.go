package main

import (
	"context"
	"os"
	"testing"

	"github.com/powerman/check"
	"github.com/powerman/gotest/testexec"
	_ "github.com/smartystreets/goconvey/convey"
)

var ctx = context.Background()

func TestMain(m *testing.M) {
	Init()
	code := m.Run()
	check.Report()
	os.Exit(code)
}

func TestFlagHelp(tt *testing.T) {
	t := check.T(tt)
	out, err := testexec.Func(ctx, t, main, "-h").CombinedOutput()
	t.Match(err, "exit status 2")
	t.Match(out, "Usage of")
}

func TestFlagVersion(tt *testing.T) {
	t := check.T(tt)
	ver = "0.0.1-test"
	out, err := testexec.Func(ctx, t, main, "-version").CombinedOutput()
	t.Nil(err)
	t.Match(out, ver)
}

func TestFlagWSHost(tt *testing.T) {
	t := check.T(tt)
	out, err := testexec.Func(ctx, t, main, "-ws.host=").CombinedOutput()
	t.Match(err, "exit status 2")
	t.Match(out, `invalid value .* -ws.host`)
}

func TestFlagWSPort(t *testing.T) {
	t.Run("Random", func(tt *testing.T) {
		t := check.T(tt)
		out, err := testexec.Func(ctx, t, main, "-ws.port=0").CombinedOutput()
		t.Match(err, "exit status 2")
		t.Match(out, `invalid value 0 .* -ws.port`)
	})
	t.Run("Negative", func(tt *testing.T) {
		t := check.T(tt)
		out, err := testexec.Func(ctx, t, main, "-ws.port=-8080").CombinedOutput()
		t.Match(err, "exit status 2")
		t.Match(out, `invalid value -8080 .* -ws.port`)
	})
}
