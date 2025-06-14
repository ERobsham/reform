package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/erobsham/reform/lib/log"
)

type ParsedLine struct {
	Timestamp time.Time   `json:"time,omitzero"`
	Host      string      `json:"host,omitempty"`
	Process   ProcessInfo `json:"proc,omitzero"`
	Message   string      `json:"msg,omitempty"`

	// optional
	LogLevel   string         `json:"level,omitempty"`
	SourceInfo SourceFileInfo `json:"src,omitzero"`
}

func ParseLine(line string) ParsedLine {
	sysTimestamp, remaining, err := ParseSystemTimeStamp(line)
	if err != nil {
		log.DebugErr("timestamp parse error", err)
	}

	hostname, remaining, err := parseHostName(remaining)
	if err != nil {
		log.DebugErr("hostname parse error", err)
	}

	proc, remaining, err := parseProcessInfo(remaining)
	if err != nil {
		log.DebugErr("process info parse error", err)
	}

	//
	// optional msg details
	//

	prefixTimestamp, remaining, err := parseMsgPrefixTimeStamp(remaining)
	prefixTimestampParsed := (err == nil)
	suffixTimestamp, remaining, err := parseMsgSuffixTimeStamp(remaining)
	suffixTimestampParsed := (err == nil)

	logLevel, remaining, err := parseLogLevel(remaining)
	logLevelParsed := (err == nil)

	sourceInfo, remaining, _ := parseSourceFileInfo(remaining)
	// sourceParsed := (err == nil)

	if !logLevelParsed {
		logLevel, remaining, _ = parseLogLevel(remaining)
	}

	if prefixTimestampParsed {
		sysTimestamp = pickMorePreciseTime(sysTimestamp, prefixTimestamp)
	}
	if suffixTimestampParsed {
		sysTimestamp = pickMorePreciseTime(sysTimestamp, suffixTimestamp)
	}

	return ParsedLine{
		Timestamp: sysTimestamp,
		Host:      hostname,
		Process:   proc,
		Message:   remaining,

		LogLevel:   logLevel,
		SourceInfo: sourceInfo,
	}
}

func (l ParsedLine) String() string {
	b := strings.Builder{}

	b.WriteString("time=" + l.Timestamp.Format(time.StampMilli) + " ")
	b.WriteString(fmt.Sprintf("host=%-10s ", trimEndTo(l.Host, 10)))
	b.WriteString(fmt.Sprintf("pid=%-6d ", l.Process.PID))
	b.WriteString(fmt.Sprintf("proc=%-10s ", trimEndTo(l.Process.Name, 10)))
	b.WriteString(fmt.Sprintf("msg=%-60s ", trimEndTo(l.Message, 60)))

	b.WriteString(fmt.Sprintf("level=%-6s ", l.LogLevel))
	b.WriteString(fmt.Sprintf("src=%-15s", trimStartTo(l.SourceInfo.Filename, 15)))
	if l.SourceInfo.LineNumber != 0 {
		b.WriteString(fmt.Sprintf(":%-5d", l.SourceInfo.LineNumber))
	}

	return b.String()
}

func trimEndTo(str string, maxLen int) string {
	strLen := len(str)
	return str[:min(strLen, maxLen)]
}
func trimStartTo(str string, maxLen int) string {
	strLen := len(str)
	startIdx := max(0, strLen-maxLen)
	return str[startIdx:strLen]
}

//#< Custom Error Type

type ParseError string

func (e ParseError) Error() string { return string(e) }

//#> Custom Error Type
