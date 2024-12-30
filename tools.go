//go:build tools
// +build tools

package main

import (
	_ "github.com/berquerant/dataclass"
	_ "github.com/go-task/task/v3/cmd/task"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "gotest.tools/gotestsum"
)
