package main

import (
	"os"

	"github.com/soli0222/misskey-summarizer/cmd/misskey-summarizer/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
