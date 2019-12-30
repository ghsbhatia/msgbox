package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghsbhatia/msgbox/pkg/msgstore"
	"github.com/ghsbhatia/msgbox/pkg/svcclient"
	"github.com/go-kit/kit/log"
)

const (
	defaultPort           = "6080"
	defaultMongoDbUrl     = "mongodb://root:secret@127.0.0.1:27017/msgbox-mongo?authSource=admin&gssapiServiceName=mongodb"
	defaultUserServiceUrl = "http://localhost:6060"
)

func main() {

	var (
		addr           = envString("PORT", defaultPort)
		httpAddr       = flag.String("http.addr", ":"+addr, "HTTP listen address")
		mongoDBUrl     = envString("MONGODB_URL", defaultMongoDbUrl)
		userServiceUrl = envString("USERSVC_URL", defaultUserServiceUrl)
		dbName         = "msgbox"
	)

	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var msgstoresvc msgstore.Service
	{
		repository, err := msgstore.NewMessageRepository(mongoDBUrl, dbName)
		if err != nil {
			logger.Log("error creating repository:", err)
			os.Exit(1)
		}
		httpclient := svcclient.NewHttpClient()
		msgstoresvc = msgstore.NewService(repository, httpclient, userServiceUrl)
	}

	httpLogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()

	mux.Handle("/", msgstore.MakeHandler(msgstoresvc, httpLogger))

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
