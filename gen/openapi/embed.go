package openapi

import "embed"

//go:embed * spark/v1/* user/v1/*
var OpenApiFs embed.FS
