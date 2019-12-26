package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghsbhatia/msgbox/pkg/useradmin"
	"github.com/go-kit/kit/log"
)

const (
	defaultPort = "8080"
)

func main() {

	var (
		addr     = envString("PORT", defaultPort)
		httpAddr = flag.String("http.addr", ":"+addr, "HTTP listen address")
	)

	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var useradminsvc useradmin.Service
	{
		repository, err := useradmin.NewUserRepository("root@tcp(127.0.0.1:3306)/msgbox")
		if err != nil {
			logger.Log("error creating repository:", err)
			os.Exit(1)
		}
		useradminsvc = useradmin.NewService(repository)
	}

	httpLogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()

	mux.Handle("/", useradmin.MakeHandler(useradminsvc, httpLogger))

	errs := make(chan error, 2)

	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServe(*httpAddr, mux)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)

}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
