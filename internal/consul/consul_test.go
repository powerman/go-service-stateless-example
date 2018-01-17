package consul

import (
	"testing"

	"github.com/powerman/check"
	_ "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) { check.TestMain(m) }

func TestResolve(tt *testing.T) {
	t := check.T(tt)

	ip, err := resolve("")
	t.Match(err, "no such host")
	t.Equal(ip, "")

	ip, err = resolve("gtld-servers.net")
	t.Match(err, "no such host")
	t.Equal(ip, "")

	ip, err = resolve("a.gtld-servers.net")
	t.Nil(err)
	t.NotEqual(ip, "")

	ip, err = resolve("localhost")
	t.Nil(err)
	t.Equal(ip, "127.0.0.1")

	ip, err = resolve("1.2.3.4")
	t.Nil(err)
	t.Equal(ip, "1.2.3.4")
}
