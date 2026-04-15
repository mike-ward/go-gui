package main

import (
	"github.com/mike-ward/go-gui/tools/requiredid"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(requiredid.Analyzer) }
