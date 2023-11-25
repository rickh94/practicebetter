package static

import (
	"embed"
	"github.com/benbjohnson/hashfs"
)

//go:embed css/* js/* dist/* fonts/*
var static embed.FS

var HashStatic = hashfs.NewFS(static)
