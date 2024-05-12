package main

import (
	"context"
	"flag"
	"github.com/charmbracelet/log"
	"github.com/zcubbs/spark/cmd/server/api"
	"github.com/zcubbs/spark/cmd/server/config"
	"github.com/zcubbs/spark/internal/utils"
	k8sJobs "github.com/zcubbs/spark/pkg/k8s/jobs"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new JobsRunner
	jobsRunner, err := k8sJobs.New(ctx,
		cfg.KubeconfigPath, cfg.MaxConcurrentJobs, cfg.QueueBufferSize, cfg.DefaultJobTimeout)
	if err != nil {
		log.Fatal("failed to create jobs runner", "error", err)
	}
	defer func(jobsRunner *k8sJobs.Runner) {
		err := jobsRunner.Shutdown()
		if err != nil {
			log.Error("failed to shutdown jobs runner", "error", err)
		}
	}(jobsRunner)

	// Create a new server
	server, err := api.NewServer(cfg, jobsRunner)
	if err != nil {
		log.Fatal("failed to create server", "error", err)
	}

	// Start the HTTP gateway
	go server.StartHttpGateway(ctx)

	// Start the Web server
	go server.StartWebServer()

	// Start the server
	server.StartGrpcServer()
}
