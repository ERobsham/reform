package main

import (
	"context"
	"errors"
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
	flag.StringVar(&a.Cmd, "cmd", "", "command to read the stdout from ie 'ssh user@host tail -F /var/log/syslog' (default: none)")
	flag.StringVar(&a.OutputPath, "out", "", "file to append processed output to -- if not set, defaults to stdout (default: none)")
	flag.StringVar(&a.ConfigPath, "config", "", "path to a json config to allow reading multiple streams at once (default: none)")

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
	outStreams := []streams.OutputStream{}

	if args.Cmd != "" {
		inStreams = append(inStreams, parseCmd(args.Cmd))
	}
	if args.OutputPath != "" {
		out, err := streams.NewOutputFile(args.OutputPath)
		if err != nil {
			log.Default().Error("invalid output path",
				slog.String("path", args.OutputPath),
				slog.String("err", err.Error()),
			)
			return
		}
		outStreams = append(outStreams, out)
	}

	if args.ConfigPath != "" {
		// TODO
	}

	if len(outStreams) == 0 {
		out := &streams.StdoutStream{}
		outStreams = append(outStreams, out)
	}

	a := streams.NewStreamAggregator(context.Background(), inStreams)
	counter := uint64(0)

	errs := []error{}
	for {
		line, err := a.Next()
		if err != nil {
			if errors.Is(err, streams.ErrStreamClosed) {
				break
			}
			errs = append(errs, err)
			break
		}

		parsed := parser.ParseLine(line)

		for _, out := range outStreams {
			err := out.Output(parsed)
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			break
		}

		counter++
	}

	for _, err := range errs {
		log.Default().Error("encountered error",
			slog.String("error", err.Error()),
		)
	}

	for _, out := range outStreams {
		out.Close()
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
