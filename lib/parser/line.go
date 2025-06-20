package parser

import (
	"time"

	"github.com/erobsham/reform/lib/log"
	"github.com/erobsham/reform/lib/types"
)

func ParseLine(line string) types.ParsedLine {
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

	if !prefixTimestampParsed {
		// some Rust loggers print the 'total runtime duration' as a prefix.
		// we're not saving it currently, but trimming it off makes further parsing easier.
		_, remaining, _ = parsePrefixDuration(remaining)
	}

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

	if sysTimestamp.Year() == 0 {
		sysTimestamp = time.Date(
			time.Now().Year(),
			sysTimestamp.Month(),
			sysTimestamp.Day(),
			sysTimestamp.Hour(),
			sysTimestamp.Minute(),
			sysTimestamp.Second(),
			sysTimestamp.Nanosecond(),
			sysTimestamp.Location(),
		)
	}

	return types.ParsedLine{
		Timestamp: sysTimestamp,
		Host:      hostname,
		Process:   proc,
		Message:   remaining,

		LogLevel:   logLevel,
		SourceInfo: sourceInfo,
	}
}

//#< Custom Error Type

type ParseError string

func (e ParseError) Error() string { return string(e) }

//#> Custom Error Type
