// Example service.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/powerman/must"
	"github.com/powerman/structlog"

	"github.com/powerman/go-service-stateless-example/internal/consul"
	"github.com/powerman/go-service-stateless-example/internal/logkey"
	"github.com/powerman/go-service-stateless-example/internal/ws"
)

var (
	app = strings.TrimSuffix(path.Base(os.Args[0]), ".test")
	ver string // set by ./build
	log = structlog.New()
	cfg struct {
		version     bool
		logLevel    string
		wsHost      string
		wsPort      int
		wsReadLimit int64
	}
)

func init() {
	hostname, _ := os.Hostname()
	flag.BoolVar(&cfg.version, "version", false, "print version")
	flag.StringVar(&cfg.logLevel, "log.level", "debug", "log `level` (debug|info|warn|err)")
	flag.StringVar(&cfg.wsHost, "ws.host", hostname, "WebSocket `host` (required)")
	flag.IntVar(&cfg.wsPort, "ws.port", 8080, "WebSocket `port` (>0)")
	flag.Int64Var(&cfg.wsReadLimit, "ws.readlimit", 4096, "WebSocket read limit (<=0 to disable)")
}

// Init provides common initialization for both app and tests.
func Init() {
	time.Local = time.UTC
	must.AbortIf = must.PanicIf

	structlog.DefaultLogger.
		AppendPrefixKeys(
			logkey.Remote,
			logkey.Func,
		).
		SetSuffixKeys(
			structlog.KeyStack,
		).
		SetKeysFormat(map[string]string{
			structlog.KeyUnit: " %6[2]s:", // set to max KeyUnit/package length
			logkey.Remote:     " %-21[2]s",
			logkey.Func:       " %[2]s:",
			logkey.HTTPHost:   " http://%[2]s",
			logkey.Host:       " %[2]s",
			logkey.Port:       ":%[2]v",
			"version":         " %s %v",
			"err":             " %s: %v",
			"json":            " %s=%#q",
		})
	log.SetDefaultKeyvals(
		structlog.KeyUnit, "main",
	)
}

func main() {
	Init()
	flag.Parse()

	// Wrong log.level is not fatal, it will be reported and set to "debug".
	switch {
	case cfg.version:
		fmt.Println(app, ver, runtime.Version())
		os.Exit(0)
	case cfg.wsHost == "":
		fatalFlagValue("required", "ws.host", cfg.wsHost)
	case cfg.wsPort <= 0: // free nginx doesn't support dynamic ports
		fatalFlagValue("must be > 0", "ws.port", cfg.wsPort)
	}

	structlog.DefaultLogger.SetLogLevel(structlog.ParseLevel(cfg.logLevel))
	log.Info("started", "version", ver)

	if err := consul.Init(); err != nil {
		log.Fatal(err)
	}

	err := consul.RegisterTCPService(app, cfg.wsHost, cfg.wsPort)
	if err != nil {
		log.Fatal(err)
	}

	if err := ws.Init(cfg.wsReadLimit); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", ws.Serve)
	log.Fatal(http.ListenAndServe(cfg.wsHost+":"+strconv.Itoa(cfg.wsPort), nil))
}

// fatalFlagValue report invalid flag values in same way as flag.Parse().
func fatalFlagValue(msg, name string, val interface{}) {
	fmt.Fprintf(os.Stderr, "invalid value %#v for flag -%s: %s\n", val, name, msg)
	flag.Usage()
	os.Exit(2)
}
