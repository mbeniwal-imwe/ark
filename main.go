package main

import (
	"os"

	"github.com/mbeniwal-imwe/ark/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
