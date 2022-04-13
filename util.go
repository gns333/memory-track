package main

import (
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/shirou/gopsutil/process"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
)

func IsRootUser() bool {
	currentUser, err := user.Current()
	if err != nil {
		PrintDebugInfo("get current user failed: %v", err)
		return false
	}
	return currentUser.Username == "root"
}

func IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err != nil {
		PrintDebugInfo("look path failed: %v", err)
		return false
	}
	return true
}

func IsRpmPackageInstalled(name string) bool {
	out, err := RunShellCommand(fmt.Sprintf("rpm -q %s", name))
	if err != nil {
		PrintDebugInfo("rpm query failed: %v", err)
		return false
	}
	return strings.HasPrefix(out, name)
}

func IsProcessRunning(pid int32) bool {
	processes, _ := process.Processes()
	for _, p := range processes {
		if p.Pid == pid {
			return true
		}
	}
	return false
}

func GetProcessExecutableFilePath(pid int32) (string, error) {
	out, err := RunShellCommand(fmt.Sprintf("ls -l /proc/%d", pid))
	if err != nil {
		return "", err
	}
	reg := regexp.MustCompile("exe ->.+")
	ret := reg.FindString(out)
	if len(ret) > 0 {
		return ret[7:], nil
	} else {
		return "", errors.New(fmt.Sprintf("find \"exe ->.+\" failed!"))
	}
}

func GetDynamicDependencyPath(execPath string, depName string) (string, error) {
	out, err := RunShellCommand(fmt.Sprintf("ldd %s", execPath))
	if err != nil {
		return "", err
	}
	regStr := fmt.Sprintf("/.+%s\\.so.+ ", depName)
	PrintDebugInfo("regexp string : \"%s\"", regStr)
	reg := regexp.MustCompile(regStr)
	ret := reg.FindString(out)
	if len(ret) > 0 {
		return ret[:len(ret)-1], nil
	} else {
		return "", errors.New(fmt.Sprintf("find \"%s\" failed!", regStr))
	}
}

func RunShellCommand(cmd string) (string, error) {
	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	PrintDebugInfo("run shell: '%s'", cmd)
	if err != nil {
		return "", err
	}
	if Debug {
		color.Debug.Prompt("=========================================")
		color.Debug.Prompt("run shell output:")
		color.Comment.Print(string(out))
		color.Debug.Prompt("=========================================")
	}
	return string(out), nil
}

func PrintVerboseInfo(format string, a ...interface{}) {
	if Verbose || Debug {
		color.Info.Prompt(format, a...)
	}
}

func PrintDebugInfo(format string, a ...interface{}) {
	if Debug {
		color.Debug.Prompt(format, a...)
	}
}
