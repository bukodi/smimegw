package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime/debug"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number and build info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n")

		bi, ok := debug.ReadBuildInfo()
		if !ok {
			panic("ReadBuildInfo failed")
		}
		fmt.Printf("Build info:\n%+v\n", bi)
	},
}
