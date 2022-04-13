package main

import (
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"strconv"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record memory stat",
	Run:   runRecordCmd,
}

var RecordPid int32
var RecordTime int32
var RecordOutPath string

func init() {
	recordCmd.Flags().Int32VarP(&RecordPid, "pid", "p", 0, "target process id")
	_ = recordCmd.MarkFlagRequired("pid")
	recordCmd.Flags().Int32VarP(&RecordTime,"time", "t", -1, "record seconds")
	recordCmd.Flags().StringVarP(&RecordOutPath,"out", "o", "", "out put file path")
	rootCmd.AddCommand(recordCmd)
}

func runRecordCmd(cmd *cobra.Command, args []string) {
	err := recordProcessMem(RecordPid)
	if err != nil {
		color.Error.Prompt("%v", err)
	}
}

func recordProcessMem(pid int32) error {
	if IsRootUser() == false {
		return errors.New("not root user")
	}
	PrintVerboseInfo("check root user [ok]")

	if IsProcessRunning(pid) == false {
		return fmt.Errorf("process id(%d) not exist", pid)
	}
	PrintVerboseInfo("check process running [ok]")

	err := checkSystemTapDependency()
	if err != nil {
		return err
	}
	PrintVerboseInfo("check systemtap dependency [ok]")

	execFilePath, err := GetProcessExecutableFilePath(pid)
	if err != nil {
		return fmt.Errorf("get exec file path error! pid(%d)\n %w", pid, err)
	}
	libstdcppPath, err := GetDynamicDependencyPath(execFilePath, "libstdc\\+\\+")
	if err != nil {
		return fmt.Errorf("get libstdc++ path error! exec(%s)\n %w", execFilePath, err)
	}
	libcPath, err := GetDynamicDependencyPath(execFilePath, "libc")
	if err != nil {
		return fmt.Errorf("get libc path error! exec(%s)\n %w", execFilePath, err)
	}
	PrintVerboseInfo("get bin paths [ok]")

	mallocStapCmd := buildMallocStapCmd(pid, execFilePath, libcPath, libstdcppPath)
	freeStapCmd := buildFreeStapCmd(pid, execFilePath, libcPath, libstdcppPath)

	color.Info.Println(mallocStapCmd)
	color.Info.Println(freeStapCmd)
	return nil
}

func checkSystemTapDependency() error {
	if IsCommandAvailable("stap") == false {
		return errors.New("require install [systemtap]")
	}
	if IsRpmPackageInstalled("glibc-debuginfo") == false {
		return errors.New("require install [glibc-debuginfo]")
	}
	if IsRpmPackageInstalled("glibc-debuginfo-common") == false {
		return errors.New("require install [glibc-debuginfo-common]")
	}
	if IsRpmPackageInstalled("glibc-devel") == false {
		return errors.New("require install [glibc-devel]")
	}
	return nil
}

func buildMallocStapCmd(pid int32, execPath string, libCPath string, libStdCppPath string) string {
	mallocCmd := "stap " +
		"-d " + libCPath +
		" -d " + libStdCppPath +
		" -d " + execPath +
		" -x " + strconv.Itoa(int(pid)) +
		" -e " +
		"'probe process(\"" + libCPath + "\").function(\"malloc\").return" +
		"{ if(pid() == target()) " +
			"{ " +
				"printf(\"" +
					"---===\\n" +
					"bytes=%d\\n" +
					"%s\\n" +
					"***===\\n" +
					"%s\\n" +
					"===***\\n" +
					"===---\\n\\n\"," +
				"@entry($bytes), $$return, sprint_ubacktrace());" +
			"} " +
		"}'"
	if Debug {
		color.Debug.Println(mallocCmd)
	}
	return mallocCmd
}

func buildFreeStapCmd(pid int32, execPath string, libCPath string, libStdCppPath string) string {
	freeCmd := "stap " +
		"-d " + libCPath +
		" -d " + libStdCppPath +
		" -d " + execPath +
		" -x " + strconv.Itoa(int(pid)) +
		" -e " +
		"'probe process(\"" + libCPath + "\").function(\"free\")" +
		"{ if(pid() == target()) " +
			"{ " +
				"printf(\"" +
					"---===\\n" +
					"mem=%d\\n" +
					"***===\\n" +
					"%s\\n" +
					"===***\\n" +
					"===---\\n\\n\"," +
				"@entry($mem), sprint_ubacktrace());" +
			"} " +
		"}'"
	if Debug {
		color.Debug.Println(freeCmd)
	}
	return freeCmd
}