package main

import (
	"fmt"
	"os"
)

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
	}
}
