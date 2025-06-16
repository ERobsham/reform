package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/erobsham/reform/lib/log"
)

type CliArgs struct {
	LogLevel   int
	ConfigPath string
	Cmd        string
	OutputPath string
	SeqServer  string
}

func ParseCmdStr(cmdStr string) (string, []string) {
	idx := strings.IndexByte(cmdStr, ' ')
	if idx == -1 {
		log.Default().Debug("init with single command stream", slog.String("cmd", cmdStr))
		return cmdStr, nil
	} else {
		cmd := cmdStr[:idx]
		args := parseCmdArgs(cmdStr[idx+1:])
		log.Default().Debug("init with single command stream",
			slog.String("cmd", cmd),
			slog.String("args", fmt.Sprintf("%v", args)))
		return cmd, args
	}
}

func parseCmdArgs(argStr string) []string {
	args := []string{}
	for {
		argLen := len(argStr)
		if len(argStr) == 0 {
			break
		}

		spaceIdx := strings.IndexByte(argStr, ' ')
		quoteIdx := strings.Index(argStr, "\"")

		// `some args "some string"`
		//      ^
		if spaceIdx != -1 && spaceIdx < quoteIdx {
			args = append(args, argStr[:spaceIdx])
			argStr = argStr[min(spaceIdx+1, argLen-1):]
			continue
		}

		// `"some string"`
		//  ^
		if quoteIdx != -1 {
			quoteIdx2 := strings.Index(argStr[1:], "\"")
			if quoteIdx2 == -1 {
				log.Default().Error("quoted string missing unquote", slog.String("remaining", argStr))
				break
			}

			args = append(args, argStr[1:quoteIdx2+1])
			argStr = argStr[min(quoteIdx2+2, argLen):]
			continue
		}

		if spaceIdx == -1 && quoteIdx == -1 {
			if argLen > 0 {
				args = append(args, argStr)
			}
			break
		}
	}
	return args
}

func ParseSeqServer(serverInfo string) (string, string) {
	idx := strings.IndexByte(serverInfo, ';')
	if idx == -1 {
		log.Default().Debug("init with seq output stream", slog.String("host", serverInfo))
		return serverInfo, ""
	} else {
		host := serverInfo[:idx]
		key := serverInfo[idx+1:]
		log.Default().Debug("init with seq output stream",
			slog.String("host", host),
			slog.String("api-key", key),
		)
		return host, key
	}
}
