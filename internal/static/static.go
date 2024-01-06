package static

import (
	"embed"
	"github.com/benbjohnson/hashfs"
)

//go:embed dist/* img/*
var static embed.FS

var HashStatic = hashfs.NewFS(static)
