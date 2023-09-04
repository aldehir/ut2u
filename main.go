package main

import (
	"os"

	"github.com/aldehir/ut2u/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
