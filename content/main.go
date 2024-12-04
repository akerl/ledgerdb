package content

import "embed"

// Static defines the embedded files
//
//go:embed static/*
var Static embed.FS
