package main

import (
	"errors"
	"github.com/gookit/color"
	"hash/crc32"
	"strconv"
	"strings"
)

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

const (
	OpStart    = "---==="
	OpEnd      = "===---"
	StackStart = "***==="
	StackEnd   = "===***"
)

func isOperationStartLine(line string) bool {
	return line == OpStart
}

func isOperationEndLine(line string) bool {
	return line == OpEnd
}

func buildMallocProbeCmdStr(pid int32, execPath string, libCPath string, libStdCppPath string) string {
	mallocCmdStr := "stap -v " +
		"-d " + libCPath +
		" -d " + libStdCppPath +
		" -d " + execPath +
		" -x " + strconv.Itoa(int(pid)) +
		" -e " +
		"'probe process(\"" + libCPath + "\").function(\"malloc\").return" +
		"{ if(pid() == target()) " +
		"{ " +
		"printf(\"" +
		OpStart + "\\n" +
		"bytes=%d\\n" +
		"%s\\n" +
		StackStart + "\\n" +
		"%s\\n" +
		StackEnd + "\\n" +
		OpEnd + "\\n\\n\"," +
		"@entry($bytes), $$return, sprint_ubacktrace());" +
		"} " +
		"}'"
	if Debug {
		color.Debug.Println(mallocCmdStr)
	}
	return mallocCmdStr
}

func buildFreeProbeCmdStr(pid int32, execPath string, libCPath string, libStdCppPath string) string {
	freeCmdStr := "stap -v " +
		"-d " + libCPath +
		" -d " + libStdCppPath +
		" -d " + execPath +
		" -x " + strconv.Itoa(int(pid)) +
		" -e " +
		"'probe process(\"" + libCPath + "\").function(\"free\")" +
		"{ if(pid() == target()) " +
		"{ " +
		"printf(\"" +
		OpStart + "\\n" +
		"mem=%d\\n" +
		StackStart + "\\n" +
		"%s\\n" +
		StackEnd + "\\n" +
		OpEnd + "\\n\\n\"," +
		"$mem, sprint_ubacktrace());" +
		"} " +
		"}'"
	if Debug {
		color.Debug.Println(freeCmdStr)
	}
	return freeCmdStr
}

func parseMallocOpStr(opStr []string) (*MallocOp, error) {
	PrintDebugInfo("###### malloc operation start ######")
	for _, s := range opStr {
		PrintDebugInfo(s)
	}

	op := &MallocOp{}
	b, err := strconv.Atoi(strings.TrimPrefix(opStr[0], "bytes="))
	if err != nil {
		return nil, err
	}
	op.byte = int64(b)
	a, err := strconv.ParseUint(strings.TrimPrefix(opStr[1], "return=0x"), 16, 64)
	if err != nil {
		return nil, err
	}
	op.addr = uintptr(a)
	op.stack = make([]string, len(opStr)-4)
	copy(op.stack, opStr[3:len(opStr)-1])
	op.stackHash = hashCodeString(op.stack)

	PrintDebugInfo("###### malloc operation parsed ######")
	PrintDebugInfo("op.byte=%d", op.byte)
	PrintDebugInfo("op.addr=%d", op.addr)
	PrintDebugInfo("op.stackhash=%d", op.stackHash)
	for _, s := range op.stack {
		PrintDebugInfo(s)
	}
	PrintDebugInfo("###### malloc operation end ######\n")
	return op, nil
}

func parseFreeOpStr(opStr []string) (*FreeOp, error) {
	PrintDebugInfo("###### free operation start ######")
	for _, s := range opStr {
		PrintDebugInfo(s)
	}

	op := &FreeOp{}
	a, err := strconv.ParseInt(strings.TrimPrefix(opStr[0], "mem="), 10, 64)
	if err != nil {
		return nil, err
	}
	op.addr = uintptr(a)
	op.stack = make([]string, len(opStr)-3)
	copy(op.stack, opStr[2:len(opStr)-1])
	op.stackHash = hashCodeString(op.stack)

	PrintDebugInfo("###### free operation parsed ######")
	PrintDebugInfo("op.addr=%d", op.addr)
	PrintDebugInfo("op.stackhash=%d", op.stackHash)
	for _, s := range op.stack {
		PrintDebugInfo(s)
	}
	PrintDebugInfo("###### free operation end ######\n")
	return op, nil
}

func hashCodeString(str []string) uint32 {
	var buf string
	for _, s := range str {
		buf += s
	}
	return crc32.ChecksumIEEE([]byte(buf))
}
