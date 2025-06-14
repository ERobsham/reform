package parser

import "strings"

const (
	ErrHostnameTooShort ParseError = "line too short to parse hostname"
)

func parseHostName(line string) (hostname string, remainder string, err error) {
	idx := strings.IndexByte(line, ' ')
	if idx == -1 {
		return "", line, ErrHostnameTooShort
	}

	hostname = line[:idx]
	if hostname == "---" {
		// apple logs messages like `--- last message repeated {n} times ---`
		// without any hostname/proc info
		return "", line, nil
	}

	return hostname, line[idx+1:], nil
}
