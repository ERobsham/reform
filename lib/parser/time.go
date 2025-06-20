package parser

import (
	"strconv"
	"time"
)

const (
	ErrTimeTooShort ParseError = "line too short to parse starting timestamp"
	ErrNotTimestamp ParseError = "no timestamp layouts matched"

	SysTimestamp_min_len = len(time.Stamp)
	SysTimestamp_max_len = len(time.Stamp + ".11223")
)

func ParseSystemTimeStamp(line string) (time.Time, string, error) {
	layouts := []string{
		time.StampMicro[:SysTimestamp_max_len],
		time.Stamp,
	}
	return parsePrefixTimestamp(line, layouts)
}

func parseMsgPrefixTimeStamp(line string) (time.Time, string, error) {
	layouts := []string{
		"15:04:05.000",
		"15:04:05",
	}
	return parsePrefixTimestamp(line, layouts)
}

func parseMsgSuffixTimeStamp(line string) (time.Time, string, error) {
	layouts := []string{
		"Jan 02 15:04:05:2006-01-02 MST",
		"Jan 2 15:04:05:2006-01-02 MST",
	}
	return parseSuffixTimestamp(line, layouts)
}

func pickMorePreciseTime(timestamp time.Time, other time.Time) time.Time {
	if timestamp.Second() != other.Second() {
		return timestamp
	}
	if timestamp.Nanosecond() == other.Nanosecond() {
		return timestamp
	}

	temp := timestamp.Truncate(time.Second)
	t1 := timestamp.Nanosecond()
	t2 := other.Nanosecond()

	for i := 10; i < int(time.Second); i *= 10 {
		r1 := t1 / i
		r2 := t2 / i

		if t1 != (r1 * i) {
			return temp.Add(time.Duration(t1))
		} else if t2 != (r2 * i) {
			return temp.Add(time.Duration(t2))
		}
	}

	return timestamp
}

func parsePrefixDuration(line string) (time.Duration, string, error) {
	lineLen := len(line)
	startIdx := consumeNextOpeningWrapper(line, 0)
	d, endIdx := consumeNextDuration(line, startIdx)

	if endIdx == startIdx || d == 0 {
		return 0, line, ErrNotTimestamp
	}

	endIdx = consumeNextClosingWrapper(line, endIdx)

	return d, line[min(endIdx, lineLen-1):], nil
}

func consumeNextDuration(line string, idx int) (d time.Duration, endIdx int) {
	endIdx = idx

	// `00:00:00.000`
	//  ^^
	hrsIdx := consumeNextNumber(line, endIdx)
	if hrsIdx == endIdx {
		return 0, idx
	}
	hrsStr := line[endIdx:hrsIdx]

	// `:00:00.000`
	//  ^
	endIdx = hrsIdx
	if endIdx+1 > len(line) || line[endIdx] != ':' {
		return 0, idx
	}
	endIdx += 1

	// `00:00.000`
	//  ^^
	minIdx := consumeNextNumber(line, endIdx)
	if minIdx == endIdx {
		return 0, idx
	}
	minsStr := line[endIdx:minIdx]

	// `:00.000`
	//  ^
	endIdx = minIdx
	if endIdx+1 > len(line) || line[endIdx] != ':' {
		return 0, idx
	}
	endIdx += 1

	// `00.000`
	//  ^^
	secIdx := consumeNextNumber(line, endIdx)
	if secIdx == endIdx {
		return 0, idx
	}
	secsStr := line[endIdx:secIdx]

	// `.000`
	//  ^
	endIdx = secIdx
	if endIdx+1 > len(line) || line[endIdx] != '.' {
		return 0, idx
	}
	endIdx += 1

	// `000`
	//  ^^^
	fracSecIdx := consumeNextNumber(line, endIdx)
	if fracSecIdx == endIdx {
		return 0, idx
	}
	fracSecStr := line[endIdx:fracSecIdx]

	hours, _ := strconv.ParseUint(hrsStr, 10, 64)
	mins, _ := strconv.ParseUint(minsStr, 10, 64)
	secs, _ := strconv.ParseFloat(secsStr+"."+fracSecStr, 64)

	d = time.Hour*time.Duration(hours) +
		time.Minute*time.Duration(mins) +
		time.Duration(float64(time.Second)*secs)
	if d > 0 {
		return d, fracSecIdx
	}

	return d, idx
}

//
// generic time parsing
//

func parsePrefixTimestamp(line string, layouts []string) (time.Time, string, error) {

	lineLen := len(line)
	startIdx := consumeNextOpeningWrapper(line, 0)

	for _, layout := range layouts {
		layoutLen := len(layout)
		if lineLen < layoutLen+1 {
			continue
		}

		timestamp, err := time.Parse(layout, line[startIdx:startIdx+layoutLen])
		if err != nil {
			continue
		}

		startIdx += layoutLen
		startIdx = consumeNextClosingWrapper(line, startIdx)

		return timestamp, line[startIdx:], nil
	}

	return time.Time{}, line, ErrNotTimestamp
}
func parseSuffixTimestamp(line string, layouts []string) (time.Time, string, error) {

	lineLen := len(line)
	lastIdx := consumePrevClosingWrapper(line, lineLen)

	for _, layout := range layouts {
		layoutLen := len(layout)
		if lineLen < layoutLen+1 {
			continue
		}
		if lastIdx-layoutLen < 0 {
			continue
		}

		timestamp, err := time.Parse(layout, line[lastIdx-layoutLen:lastIdx])
		if err != nil {
			continue
		}

		lastIdx -= layoutLen
		lastIdx = consumePrevOpeningWrapper(line, lastIdx)
		return timestamp, line[:lastIdx], nil
	}

	return time.Time{}, line, ErrNotTimestamp
}
