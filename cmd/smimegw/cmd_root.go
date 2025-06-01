package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "smimegw",
	Short: "SMTP MTA for decrypting S/MIME messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no sub command is given, use serve as default
		return cmd.Help()
	},
}
