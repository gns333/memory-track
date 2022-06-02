package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"os/exec"
	"strings"
	"time"
)

var stopRecord = make(chan bool, 1)
var mallocStatMap = make(map[uint32]*MallocStat)
var freeStatMap = make(map[uint32]*FreeStat)
var remainMallocOpMap = make(map[uintptr]*MallocOp)

type MallocStat struct {
	Count int32
	Byte  int64
	Stack []string
}

type FreeStat struct {
	Count int32
	Stack []string
}

type MallocOp struct {
	Byte      int64
	Addr      uintptr
	Stack     []string
	StackHash uint32
}

type FreeOp struct {
	Addr      uintptr
	Stack     []string
	StackHash uint32
}

func RecordProcessMem(pid int32) error {
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

	mc := make(chan *MallocOp, 100)
	fc := make(chan *FreeOp, 100)
	ec := make(chan error, 100)
	probeMemoryOperation(pid, mc, fc, ec)

	color.Info.Prompt("start track memory...")
	color.Info.Prompt("press [ctrl + C] stop")

	setupStopTimer()
	
Loop:
	for {
		select {
		case err := <-ec:
			PrintVerboseInfo("probe: %v", err)
		case free := <-fc:
			addFreeOp(free)
		case malloc := <-mc:
			addMallocOp(malloc)
		case <-stopRecord:
			break Loop
		default:
		}
		time.Sleep(time.Millisecond)
	}

	savePath, err := Save()
	if err != nil {
		return err
	}
	color.Info.Prompt("save data to [%s]", savePath)

	return nil
}

func StopRecordMem() {
	stopRecord <- true
}

func setupStopTimer() {
	if RecordTime > 0 {
		go func() {
			timeTicker := time.NewTicker(time.Second)
			leftSecond := RecordTime
			for {
				select {
				case <-timeTicker.C:
					leftSecond--
					color.Info.Prompt("finish after %d second...", leftSecond)
					if leftSecond <= 0 {
						StopRecordMem()
					}
				}
			}
		}()
	}
}

func addMallocOp(m *MallocOp) {
	if _, ok := mallocStatMap[m.StackHash]; ok {
		mallocStatMap[m.StackHash].Count += 1
		mallocStatMap[m.StackHash].Byte += m.Byte
	} else {
		mallocStatMap[m.StackHash] = &MallocStat{
			Byte:  m.Byte,
			Count: 1,
			Stack: m.Stack,
		}
	}
	remainMallocOpMap[m.Addr] = m
}

func addFreeOp(f *FreeOp) {
	if _, ok := freeStatMap[f.StackHash]; ok {
		freeStatMap[f.StackHash].Count += 1
	} else {
		freeStatMap[f.StackHash] = &FreeStat{
			Count: 1,
			Stack: f.Stack,
		}
	}
	delete(remainMallocOpMap, f.Addr)
}

func probeMemoryOperation(pid int32, mc chan *MallocOp, fc chan *FreeOp, ec chan error) {
	execFilePath, libstdcppPath, libcPath, err := getBinFilePath(pid)
	if err != nil {
		ec <- err
		return
	}

	mallocCmdStr := buildMallocProbeCmdStr(pid, execFilePath, libcPath, libstdcppPath)
	freeCmdStr := buildFreeProbeCmdStr(pid, execFilePath, libcPath, libstdcppPath)

	mallocProbeCommand := exec.Command("/bin/sh", "-c", mallocCmdStr)
	freeProbeCommand := exec.Command("/bin/sh", "-c", freeCmdStr)

	mallocOutReader, mallocErrReader, err := getStdPipeReader(mallocProbeCommand)
	if err != nil {
		ec <- fmt.Errorf("get malloc pipe reader: %w", err)
		return
	}

	freeOutReader, freeErrReader, err := getStdPipeReader(freeProbeCommand)
	if err != nil {
		ec <- fmt.Errorf("get free pipe reader: %w", err)
		return
	}

	err = mallocProbeCommand.Start()
	if err != nil {
		ec <- fmt.Errorf("malloc cmd start error: %w", err)
		return
	}

	err = freeProbeCommand.Start()
	if err != nil {
		ec <- fmt.Errorf("free cmd start error: %w", err)
		return
	}

	go checkErrReader(mallocErrReader, ec)
	go checkErrReader(freeErrReader, ec)
	go collectMallocOp(mallocOutReader, mc, ec)
	go collectFreeOp(freeOutReader, fc, ec)
}

func checkErrReader(buf *bufio.Reader, ec chan error) {
	for {
		output, _, err := buf.ReadLine()
		if err == nil {
			if strings.Index(string(output), "Missing separate debuginfos") < 0 {
				ec <- fmt.Errorf("std err out put: %s", output)
			}
		} else {
			if err.Error() != "EOF" {
				ec <- fmt.Errorf("std err error: %w", err)
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func getStdPipeReader(command *exec.Cmd) (*bufio.Reader, *bufio.Reader, error) {
	stdOutPipe, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stdOutReader := bufio.NewReader(stdOutPipe)

	stdErrPipe, err := command.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	stdErrReader := bufio.NewReader(stdErrPipe)

	return stdOutReader, stdErrReader, nil
}

func getBinFilePath(pid int32) (string, string, string, error) {
	execFilePath, err := GetProcessExecutableFilePath(pid)
	if err != nil {
		return "", "", "", fmt.Errorf("get exec file path error! pid(%d)\n %w", pid, err)
	}
	libcPath, err := GetDynamicDependencyPath(execFilePath, "libc")
	if err != nil {
		return "", "", "", fmt.Errorf("get libc path error! exec(%s)\n %w", execFilePath, err)
	}
	libstdcppPath, err := GetDynamicDependencyPath(execFilePath, "libstdc\\+\\+")
	if err != nil {
		PrintDebugInfo("get libstdc++ path faild! exec(%s)\n %w", execFilePath, err)
	}
	return execFilePath, libstdcppPath, libcPath, nil
}

func collectMallocOp(mallocOutReader *bufio.Reader, mc chan *MallocOp, ec chan error) {
	var mallocOpBuf []string
	isMallocOpRange := false
	for {
		output, _, err := mallocOutReader.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				ec <- fmt.Errorf("malloc probe std out error: %w", err)
			}
		} else {
			if isOperationStartLine(string(output)) {
				isMallocOpRange = true
			} else if isOperationEndLine(string(output)) {
				isMallocOpRange = false
				op, err := parseMallocOpStr(mallocOpBuf)
				if err != nil {
					ec <- fmt.Errorf("parse malloc op str error: %w", err)
				} else {
					mc <- op
				}
				mallocOpBuf = mallocOpBuf[:0]
			} else {
				if isMallocOpRange {
					mallocOpBuf = append(mallocOpBuf, string(output))
				}
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func collectFreeOp(freeOutReader *bufio.Reader, fc chan *FreeOp, ec chan error) {
	var freeOpBuf []string
	isFreeOpRange := false
	for {
		output, _, err := freeOutReader.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				ec <- fmt.Errorf("free probe std out error: %w", err)
			}
		} else {
			if isOperationStartLine(string(output)) {
				isFreeOpRange = true
			} else if isOperationEndLine(string(output)) {
				isFreeOpRange = false
				op, err := parseFreeOpStr(freeOpBuf)
				if err != nil {
					ec <- fmt.Errorf("parse free op str error: %w", err)
					return
				} else {
					fc <- op
				}
				freeOpBuf = freeOpBuf[:0]
			} else {
				if isFreeOpRange {
					freeOpBuf = append(freeOpBuf, string(output))
				}
			}
		}
		time.Sleep(time.Millisecond)
	}
}
