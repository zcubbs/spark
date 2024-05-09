package config

import "time"

type Configuration struct {
	Debug             bool             `mapstructure:"debug"`
	DevMode           bool             `mapstructure:"dev_mode"`
	HttpServer        HttpServerConfig `mapstructure:"http_server"`
	GrpcServer        GrpcServerConfig `mapstructure:"grpc_server"`
	Auth              AuthConfig       `mapstructure:"auth"`
	InitAdminPassword string           `mapstructure:"init_admin_password"`
	// KubeconfigPath is the path to the kubeconfig file.
	// If empty, the in-cluster config will be used.
	// Default value: ""
	// If value is "default", the default kubeconfig file will be used.
	// otherwise, the value is the path to the kubeconfig file.
	KubeconfigPath string `mapstructure:"kubeconfig_path"`
	// MaxConcurrentJobs is the maximum number of concurrent jobs that can be run at the same time.
	MaxConcurrentJobs int `mapstructure:"max_concurrent_jobs"`
	// RateLimitRequestsPerSecond is the maximum number of requests per second allowed.
	RateLimitRequestsPerSecond int `mapstructure:"rate_limit_requests_per_second"`
	// RateLimitBurst is the maximum number of requests that can be burst(ed).
	RateLimitBurst int `mapstructure:"rate_limit_burst"`

	// Version is the version of the application.
	Version string `json:"version"`
	// Commit is the git commit of the application.
	Commit string `json:"commit"`
	// Date is the build date of the application.
	Date string `json:"date"`
}

type HttpServerConfig struct {
	Port         int    `mapstructure:"port"`
	AllowOrigins string `mapstructure:"allow_origins"`
	AllowHeaders string `mapstructure:"allow_headers"`
	TZ           string `mapstructure:"tz"`
	// ReadHeaderTimeout is the amount of time allowed to read request headers. Default values: '3s'
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
}

type GrpcServerConfig struct {
	Port             int       `mapstructure:"port"`
	EnableReflection bool      `mapstructure:"enable_reflection"`
	Tls              TlsConfig `mapstructure:"tls"`
}

type TlsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Cert    string `mapstructure:"cert"`
	Key     string `mapstructure:"key"`
}

type AuthConfig struct {
	TokenSymmetricKey    string        `mapstructure:"token_symmetric_key"`
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
}
