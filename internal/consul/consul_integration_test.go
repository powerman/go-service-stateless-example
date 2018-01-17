// +build integration

package consul_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/powerman/check"

	"github.com/powerman/go-service-stateless-example/internal/consul"
)

func init() {
	if err := consul.Init(); err != nil {
		panic(err)
	}
}

func TestRegisterTCPService(tt *testing.T) {
	t := check.T(tt)

	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	t.Match(consul.RegisterTCPService("", "", 0), "no such host")
	t.Match(consul.RegisterTCPService("", "localhost", 0), "service name")
	t.Nil(consul.RegisterTCPService("test", "localhost", 0))
}
