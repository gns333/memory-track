package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report memory stat",
	Run:   runReportCmd,
	Args:  cobra.ExactValidArgs(2),
}

func init() {
	rootCmd.AddCommand(reportCmd)
}

func runReportCmd(cmd *cobra.Command, args []string) {
	fmt.Println(args)
}