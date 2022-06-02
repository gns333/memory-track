package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:     "memory-track",
	Short:   "A memory track tool",
	Example: "memory-track record -p pid [-t sec] [-o path]\nmemory-track report -i path",
}

var Verbose bool
var Debug bool

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "print verbose log")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "print debug log")
	rootCmd.PersistentFlags().MarkHidden("debug")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
