/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	TAG     string
	GOVER   string
	COMMIT  string
	BLDTIME string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version info",
	Run: func(cmd *cobra.Command, args []string) {
		ver := "web3-console-backend\n" +
			"  Version: " + TAG + "\n" +
			"  Commit ID: " + COMMIT + "\n" +
			"  Build: " + BLDTIME + "\n" +
			"  Go Version: " + GOVER + "\n"
		fmt.Print(ver)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
