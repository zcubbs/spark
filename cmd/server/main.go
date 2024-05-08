package main

import (
	"flag"
	"github.com/charmbracelet/log"
	"github.com/zcubbs/spark/cmd/server/api"
	"github.com/zcubbs/spark/cmd/server/config"
	"github.com/zcubbs/spark/internal/utils"
	"github.com/zcubbs/x/pretty"
	"os"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var cfg *config.Configuration

var configPath = flag.String("config", "", "Path to the configuration file")

func init() {
	flag.Parse()

	// Load configuration
	log.Info("loading configuration...")
	var err error
	err = utils.Load(*configPath, &cfg, config.Defaults, config.EnvKeys)
	if err != nil {
		log.Fatal("failed to load configuration", "error", err)
	}

	cfg.Version = Version
	cfg.Commit = Commit
	cfg.Date = Date

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
		pretty.PrintJson(cfg)
	}

	if !cfg.DevMode {
		log.SetFormatter(log.JSONFormatter)
	}

	// Set the timezone
	err = os.Setenv("TZ", cfg.HttpServer.TZ)
	if err != nil {
		log.Error("failed to set timezone", "error", err)
	}
	utils.CheckTimeZone()

	log.Info("loaded configuration")
}

func main() {
	// Create a new server
	server, err := api.NewServer(cfg)
	if err != nil {
		log.Fatal("failed to create server", "error", err)
	}

	// Start the server
	go server.StartGrpcServer()

	// Start the HTTP gateway
	server.StartHttpGateway()
}
