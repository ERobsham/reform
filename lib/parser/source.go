package parser

import (
	"strconv"
	"strings"

	"github.com/erobsham/reform/lib/types"
)

func parseSourceFileInfo(line string) (sfInfo types.SourceFileInfo, remainder string, err error) {
	suffixTypeMap := map[string]string{
		"C":           ".c",
		"C++":         ".cpp",
		"Objective-C": ".m",
		"C#":          ".cs",
		"Java":        ".java",
		"Swift":       ".swift",
		"Go":          ".go",
		"Rust":        ".rs",
	}

	trimmed := strings.TrimLeft(line, " ")
	sfInfo, remainder, err = handleSpecialCases(trimmed)
	if err == nil {
		return
	}

	for lang, suffix := range suffixTypeMap {
		endIdx := strings.Index(trimmed, suffix)
		if endIdx == -1 {
			continue
		}

		endIdx += len(suffix)

		if len(trimmed) > endIdx+1 && !isInSet(trimmed[endIdx], stdWrapperOrSrcEndingSuffixMap) {
			continue
		}

		sfInfo.Language = lang

		startIdx := startIdxOfFilepath(trimmed[:endIdx])

		isPrefix := startIdx <= 1 // maybe started with a 'wrapper' char
		sfInfo.Filename = trimmed[startIdx:endIdx]

		sfInfo.LineNumber, endIdx = consumeLineNumSuffix(trimmed, endIdx)
		endIdx = consumeNextClosingWrapper(trimmed, endIdx)

		if isPrefix {
			remainder = trimmed[endIdx:]
			break
		}

		startIdx = consumeCommonFilePrefixes(trimmed, startIdx)
		startIdx = consumePrevOpeningWrapper(trimmed, startIdx)
		remainder = trimmed[:startIdx]
		if sfInfo.LineNumber != 0 {
			break
		}

		if lineNum, newStartIdx := consumeLabeledLineNumSuffix(trimmed, startIdx); lineNum != 0 {
			sfInfo.LineNumber = lineNum
			remainder = trimmed[:newStartIdx]
		} else if lineNum, _ := consumeLabeledLineNumSuffix(trimmed, len(trimmed)); lineNum != 0 {
			sfInfo.LineNumber = lineNum
		}
		break
	}

	return sfInfo, remainder, nil
}

func handleSpecialCases(line string) (sfInfo types.SourceFileInfo, remainder string, err error) {

	// handle special case seen in some rust programs: like
	// `2:03:04.567 ::: cool_crate::useful_module [WARN] something went wrong!`
	//              ^^^^
	if len(line) > 4 && strings.HasPrefix(line, "::: ") {
		line = line[4:]

		crate, idx := consumeNextRustCrateName(line, 0)
		sfInfo.Filename = crate

		return sfInfo, line[idx:], nil
	}

	return sfInfo, line, ParseError("not special case")
}

func startIdxOfFilepath(chunk string) int {
	// `<line:001459 file:/src/common/engine/ipc/cool_ipc.c`>
	//                    ^--------------------------------
	// `Main.swift:40`
	//  ^------------------
	// `<socket to some-service did close> <line:001174 file:src/srvMan.m>`
	//                                                       ^-----------
	idx := consumePrevPathChars(chunk, len(chunk))
	return idx
}

func consumeLineNumSuffix(line string, endIdx int) (lineNum uint64, idx int) {
	if len(line) > endIdx+2 && line[endIdx] == ':' {
		remaining := line[endIdx+1:]

		numEndIdx := consumeNextNumber(remaining, 0)
		if numEndIdx > 0 {
			lineNum, err := strconv.ParseUint(remaining[:numEndIdx], 10, 64)
			if err != nil {
				return 0, endIdx
			}

			return lineNum, endIdx + numEndIdx + 1
		}
	}
	return lineNum, endIdx
}

func consumeLabeledLineNumSuffix(line string, endIdx int) (lineNum uint64, idx int) {
	remaining := line[:endIdx]
	lastSpaceIdx := strings.LastIndexByte(remaining, ' ')
	if lastSpaceIdx == -1 {
		return 0, endIdx
	}
	if lastSpaceIdx-1 > 0 && remaining[lastSpaceIdx-1] == ':' {
		// we might have a `line: {num}` situation
		lastSpaceIdx2 := strings.LastIndexByte(remaining[:lastSpaceIdx-1], ' ')
		if lastSpaceIdx2 != -1 {
			lastSpaceIdx = lastSpaceIdx2
		}
	}

	remaining = remaining[lastSpaceIdx:]
	remaining = strings.TrimLeft(remaining, stdWrapperPrefixes)

	stdLinePrefixes := map[string]struct{}{
		"line: ": {},
		"line:":  {},
	}
	remaining = trimSuffixSet(remaining, stdLinePrefixes)

	remaining = strings.TrimRight(remaining, stdWrapperSuffixes)

	lineNum, _ = strconv.ParseUint(remaining, 10, 64)
	if lineNum != 0 {
		endIdx = lastSpaceIdx
	}

	return lineNum, endIdx
}

func trimSuffixSet(str string, suffixSet map[string]struct{}) string {
	strLen := len(str)
	for suffix := range suffixSet {
		suffixLen := len(suffix)
		if suffixLen > strLen {
			continue
		}

		if _, exists := suffixSet[str[:suffixLen]]; exists {
			return str[suffixLen:]
		}
	}
	return str
}

func consumeCommonFilePrefixes(line string, startIdx int) int {
	stdFilePrefixes := map[string]struct{}{
		"file:":    {},
		"file: ":   {},
		"source:":  {},
		"source: ": {},
	}

	for prefix := range stdFilePrefixes {
		prefixOffset := startIdx - len(prefix)
		if prefixOffset < 0 {
			continue
		}

		p := line[prefixOffset:startIdx]
		if _, exists := stdFilePrefixes[p]; exists {
			startIdx = prefixOffset
			break
		}
	}

	return startIdx
}

func consumeNextRustCrateName(line string, startIdx int) (crate string, idx int) {
	idx = startIdx
	for {
		nextIdx := consumeNext(line, idx, isValidRustCrateChar)
		if idx == nextIdx {
			break
		}

		idx = nextIdx

		if len(line[nextIdx:]) > 2 && line[nextIdx] == ':' && line[nextIdx+1] == ':' {
			idx += 2
		}
	}
	return line[startIdx:idx], idx
}
