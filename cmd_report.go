package main

import (
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Report memory statistics by malloc usage",
	Run:   runReportCmd,
	//Args:  cobra.ExactValidArgs(2),
}

var ReportInputPath string
var ReportMinByte int64
var ReportMinCount int32

func init() {
	reportCmd.Flags().StringVarP(&ReportInputPath, "input", "i", "", "input file path")
	reportCmd.Flags().Int64VarP(&ReportMinByte, "min_byte", "b", 100, "greater than the specified byte is displayed")
	reportCmd.Flags().Int32VarP(&ReportMinCount, "min_count", "c", 10, "greater than the specified count is displayed")
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
