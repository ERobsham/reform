package types

import (
	"fmt"
	"strings"
	"time"
)

// special keys from '[CLEF](https://clef-json.org/)' standard:
// `@t` -- timestamp
// `@m` -- message
// `@l` -- log level

type ParsedLine struct {
	Timestamp time.Time   `json:"@t,omitzero"`
	Host      string      `json:"host,omitempty"`
	Process   ProcessInfo `json:"proc,omitzero"`
	Message   string      `json:"@m,omitempty"`

	// optional
	LogLevel   string         `json:"@l,omitempty"`
	SourceInfo SourceFileInfo `json:"src,omitzero"`
}

type ProcessInfo struct {
	Name string
	PID  uint64
	TID  uint64
}

type SourceFileInfo struct {
	Language   string `json:"lang,omitempty"`
	Filename   string `json:"file,omitempty"`
	LineNumber uint64 `json:"line,omitempty"`
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
