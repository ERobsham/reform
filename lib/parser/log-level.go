package parser

import "strings"

func parseLogLevel(line string) (logLevel string, remainder string, err error) {
	const max_prefix_len = 9 + 2 // (max keylen + 2 for any 'wrappers')
	stdLogLevels := map[string]struct{}{
		"trace":     {},
		"dbg":       {},
		"debug":     {},
		"debugging": {},
		"inf":       {},
		"info":      {},
		"notice":    {},
		"wrn":       {},
		"warn":      {},
		"warning":   {},
		"err":       {},
		"error":     {},
		"crit":      {},
		"critical":  {},
		"alert":     {},
		"emerg":     {},
		"emergency": {},
	}
	levelNormalizationMap := map[string]string{
		"trace":     "debug",
		"dbg":       "debug",
		"debug":     "debug",
		"debugging": "debug",
		"inf":       "info",
		"info":      "info",
		"notice":    "info",
		"warn":      "warn",
		"wrn":       "warn",
		"warning":   "warn",
		"err":       "error",
		"error":     "error",
		"crit":      "crit",
		"critical":  "crit",
		"alert":     "alert",
		"emerg":     "alert",
		"emergency": "alert",
	}

	line = strings.TrimSpace(line)
	if len(line) < max_prefix_len+1 {
		return "", line, ParseError("too short for LogLevel Parsing")
	}

	// common 'log level' wrappers `[info]` | `(info)` | `<info>` | `{info}` | ` info `
	idx := strings.IndexAny(line, stdWrapperSuffixes)

	if idx == -1 || idx > max_prefix_len {
		return "", line, ParseError("LogLevel prefix not found")
	}

	prefix := line[:idx]
	prefix = strings.Trim(prefix, stdWrappers)
	prefix = strings.ToLower(prefix)

	idx = consumeNextSpace(line, idx)

	if _, exists := stdLogLevels[prefix]; exists {
		return levelNormalizationMap[prefix], line[idx+1:], nil
	} else {
		return "", line, ParseError("LogLevel prefix not found")
	}
}
