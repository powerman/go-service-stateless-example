// Package consul provide helpers for interacting with consul.
//
// Use defaults to connect to consul ($CONSUL_HTTP_ADDR or localhost).
//
// TODO If service can be somehow unregistered (consul has restarted and
// lost this registration - shouldn't happens in production where consul
// is running with persistence, but may happens in -dev) then detect this
// and re-register or crash.
package consul

import (
	"net"
	"strconv"

	"github.com/hashicorp/consul/api"
	"github.com/powerman/structlog"

	"github.com/powerman/go-service-stateless-example/internal/logkey"
)

var log = structlog.New()

var client *api.Client

// Init must be called once before using this package.
func Init() (err error) {
	client, err = api.NewClient(api.DefaultConfig())
	return err
}

// RegisterTCPService register service in consul with TCP check.
func RegisterTCPService(name, addr string, port int) error {
	const (
		checkDeregister = "3s"
		checkInterval   = "1s"
		checkTimeout    = "1s"
	)

	// We need to register using IP and not hostname, because otherwise
	// consul DNS will reply with CNAME+A while nginx expects A only.
	ip, err := resolve(addr)
	if err != nil {
		return log.Err("failed to resolve", "err", err)
	}

	err = client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		Name:    name,
		Port:    port,
		Address: ip,
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: checkDeregister,
			Interval:                       checkInterval,
			Timeout:                        checkTimeout,
			TCP:                            ip + ":" + strconv.Itoa(port),
		},
	})
	if err != nil {
		return log.Err("failed to register service", "err", err)
	}
	log.Info("service registered", "name", name, logkey.Host, ip, logkey.Port, port)
	return nil
}

func resolve(addr string) (string, error) {
	ips, err := net.LookupIP(addr)
	if err != nil {
		return "", err
	}
	return ips[0].String(), nil
}
