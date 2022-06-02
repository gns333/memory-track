package main

import (
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record target process malloc/free call",
	Run:   runRecordCmd,
}

var RecordPid int32
var RecordTime int32
var RecordOutPath string

func init() {
	recordCmd.Flags().Int32VarP(&RecordPid, "pid", "p", 0, "target process id")
	_ = recordCmd.MarkFlagRequired("pid")
	recordCmd.Flags().Int32VarP(&RecordTime, "time", "t", -1, "record seconds")
	recordCmd.Flags().StringVarP(&RecordOutPath, "output", "o", "", "output file path")
	rootCmd.AddCommand(recordCmd)
}

func runRecordCmd(cmd *cobra.Command, args []string) {
	err := RecordProcessMem(RecordPid)
	if err != nil {
		color.Error.Prompt("%v", err)
	}
}
