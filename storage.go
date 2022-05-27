package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

type storageData struct {
	MSMap map[uint32]*MallocStat
	FSMap map[uint32]*FreeStat
	MOMap map[uintptr]*MallocOp
}

func Save() error {
	saveFilePath := RecordOutPath
	if len(saveFilePath) == 0 {
		saveFilePath = fmt.Sprintf("%d-%s.track", RecordPid, time.Now().Format("20060102150405"))
	}

	saveFile, err := os.OpenFile(saveFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}
	defer saveFile.Close()

	data := storageData{}
	data.MSMap = mallocStatMap
	data.FSMap = freeStatMap
	data.MOMap = remainMallocOpMap

	gobEncoder := gob.NewEncoder(saveFile)
	err = gobEncoder.Encode(data)
	if err != nil {
		return fmt.Errorf("gob encode error: %v", err)
	}

	return nil
}

func Load(filename string) error {
	loadFile, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}
	defer loadFile.Close()

	data := storageData{}
	gobDecoder := gob.NewDecoder(loadFile)
	err = gobDecoder.Decode(&data)
	if err != nil {
		return fmt.Errorf("gob decode error: %v", err)
	}

	mallocStatMap = data.MSMap
	freeStatMap = data.FSMap
	remainMallocOpMap = data.MOMap

	return nil
}
