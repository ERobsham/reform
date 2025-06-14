package parser

import "time"

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
