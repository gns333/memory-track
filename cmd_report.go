package main

import (
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report memory stat",
	Run:   runReportCmd,
	//Args:  cobra.ExactValidArgs(2),
}

var ReportInputPath string

func init() {
	reportCmd.Flags().StringVarP(&ReportInputPath, "input", "i", "", "input file path")
	_ = reportCmd.MarkFlagRequired("input")
	rootCmd.AddCommand(reportCmd)
}

func runReportCmd(cmd *cobra.Command, args []string) {
	err := Load(ReportInputPath)
	if err != nil {
		color.Error.Prompt("%v", err)
	}
	err = ShowReportUI()
	if err != nil {
		color.Error.Prompt("%v", err)
	}
}
