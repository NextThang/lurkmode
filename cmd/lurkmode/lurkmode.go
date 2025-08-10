package main

import (
	"fmt"
	"os"

	"github.com/nextthang/lurkmode/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: lurkmode <channel_name>")
		os.Exit(1)
	}
	if err := app.Run(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
