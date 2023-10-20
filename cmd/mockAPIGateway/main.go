package main

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/rpc"
	"net/url"
)

type config struct {
	host        string
	port        int
	contextPath string
}

type application struct {
	logger    *zap.Logger
	config    config
	rpcClient *rpc.Client
}

func main() {
	logger := zap.Must(zap.NewDevelopment())

	viper.SetEnvPrefix("MOCK_GATEWAY")
	viper.SetDefault("host", "localhost")
	viper.MustBindEnv("host")
	viper.SetDefault("port", "8080")
	viper.MustBindEnv("port")
	viper.SetDefault("context_path", "/")
	viper.MustBindEnv("context_path")

	// Ensure we have a valid URL and only take the path
	cp, err := url.Parse(viper.GetString("context_path"))
	if err != nil {
		logger.Fatal("binding contextPath", zap.Error(err))
	}

	client, err := rpc.Dial("tcp", "localhost:3000")
	if err != nil {
		logger.Fatal("creating RPC client", zap.Error(err))
	}

	app := &application{
		logger: logger,
		config: config{
			host:        viper.GetString("host"),
			port:        viper.GetInt("port"),
			contextPath: cp.Path,
		},
		rpcClient: client,
	}

	app.serve()
}
