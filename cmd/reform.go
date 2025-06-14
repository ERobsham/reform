package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"strings"

	"github.com/erobsham/reform/lib/log"
	"github.com/erobsham/reform/lib/parser"
	"github.com/erobsham/reform/lib/streams"
)

type cliArgs struct {
	LogLevel   int
	ConfigPath string
	Cmd        string
	OutputPath string
}

func parseArgs() cliArgs {
	a := cliArgs{}

	flag.IntVar(&a.LogLevel, "log", int(slog.LevelInfo), "set log level (default: 0)")
	flag.StringVar(&a.ConfigPath, "config", "", "path to a json config to allow reading multiple streams at once (default: none)")
	flag.StringVar(&a.Cmd, "cmd", "", "command to read the stdout from ie 'ssh user@host tail -F /var/log/syslog' (default: none)")
	flag.StringVar(&a.Cmd, "out", "", "file to append processed output to -- if not set, defaults to stdout (default: none)")

	flag.Parse()

	return a
}

func main() {
	args := parseArgs()

	// clamp loglevel to the range we're using
	logLevel := max(slog.Level(args.LogLevel), slog.LevelDebug)
	logLevel = min(logLevel, slog.LevelError)
	log.SetDefaultLogLevel(logLevel)

	inStreams := []streams.InputStream{}
	if args.Cmd != "" {
		inStreams = append(inStreams, parseCmd(args.Cmd))
	}
	if args.ConfigPath != "" {
		// TODO
	}
	if args.OutputPath != "" {
		// TODO
	}

	a := streams.NewStreamAggregator(context.Background(), inStreams)
	counter := uint64(0)

	for {
		line, err := a.Next()
		if err != nil {
			break
		}

		parsed := parser.ParseLine(line)
		fmt.Printf("%05d: %s \n", counter, parsed)

		counter++
	}

	a.Close()
}

func parseCmd(cmdStr string) streams.InputStream {
	idx := strings.IndexByte(cmdStr, ' ')
	if idx == -1 {
		log.Default().Debug("init with single command stream", slog.String("cmd", cmdStr))

		s := streams.NewCmdStream(context.Background(), "cmd", []string{cmdStr})
		return s
	} else {
		cmd := cmdStr[:idx]
		args := parseCmdArgs(cmdStr[idx+1:])
		log.Default().Debug("init with single command stream",
			slog.String("cmd", cmd),
			slog.String("args", fmt.Sprintf("%v", args)))

		cmdArgs := []string{cmd}
		cmdArgs = append(cmdArgs, args...)
		s := streams.NewCmdStream(context.Background(), "cmd", cmdArgs)
		return s
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
