package main

import (
	"fmt"
	"os"

	"github.com/aldehir/ut2u/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
