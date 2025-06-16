package parser

import (
	"strconv"
	"strings"

	"github.com/erobsham/reform/lib/types"
)

func parseProcessInfo(line string) (proc types.ProcessInfo, remainder string, err error) {
	// `Process Name[pid]: ... message ...`
	// `process[pid][tid]: ... message ...`

	idx := strings.IndexByte(line, ':')
	if idx == -1 {
		return types.ProcessInfo{}, line, ParseError("no ':' found before EoL")
	}
	procInfo := line[:idx]

	name, idInfo, err := parseProcName(procInfo)
	if err != nil {
		return types.ProcessInfo{}, line, err
	}

	pid, tid, err := parseProcIDs(idInfo)
	if err != nil {
		return types.ProcessInfo{}, line, err
	}

	idx = consumeNextSpace(line, idx)

	return types.ProcessInfo{
		Name: name,
		PID:  pid,
		TID:  tid,
	}, line[idx+1:], nil
}

func parseProcName(procInfo string) (name string, remainder string, err error) {
	// `Process Name[pid]`
	// `process[pid][tid]`
	idx := strings.IndexByte(procInfo, '[')
	if idx == -1 {
		return "", procInfo, ParseError("no '[' found")
	}

	return procInfo[:idx], procInfo[idx:], nil
}

func parseProcIDs(idInfo string) (pid uint64, tid uint64, err error) {
	// `[pid]`
	// `[pid][tid]`
	if len(idInfo) < 3 {
		return 0, 0, ParseError("too short to parse pid/tid info")
	}

	idInfo = idInfo[1 : len(idInfo)-1]

	idx := strings.IndexByte(idInfo, ']')
	if idx == -1 {
		pid, _ = strconv.ParseUint(idInfo, 10, 64)
	} else {
		pidStr := idInfo[:idx]
		tidStr := idInfo[idx+2:]

		pid, _ = strconv.ParseUint(pidStr, 10, 64)
		tid, _ = strconv.ParseUint(tidStr, 10, 64)
	}

	return
}
