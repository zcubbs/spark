package web

import "embed"

//go:embed static/*
var FsStaticFiles embed.FS

//go:embed templates/*
var FsTemplates embed.FS
