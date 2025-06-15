package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"

	"github.com/erobsham/reform/lib/config"
	"github.com/erobsham/reform/lib/log"
	"github.com/erobsham/reform/lib/parser"
	"github.com/erobsham/reform/lib/streams"
)

func parseArgs() config.CliArgs {
	a := config.CliArgs{}

	flag.IntVar(&a.LogLevel, "log", int(slog.LevelInfo), "set log level (default: 0)")
	flag.StringVar(&a.Cmd, "cmd", "", "command to read the stdout from ie 'ssh user@host tail -F /var/log/syslog' (default: none)")
	flag.StringVar(&a.OutputPath, "out", "", "file to append processed output to -- if not set, defaults to stdout (default: none)")
	flag.StringVar(&a.ConfigPath, "config", "", "path to a json config to allow reading multiple streams at once (default: none)")

	flag.Parse()

	return a
}

func main() {
	args := parseArgs()

	inStreams, outStreams := handleArgs(args)
	if len(inStreams) == 0 || len(outStreams) == 0 {
		return
	}

	runloop(inStreams, outStreams)
}

func handleArgs(args config.CliArgs) (inStreams []streams.InputStream, outStreams []streams.OutputStream) {
	inStreams = []streams.InputStream{}
	outStreams = []streams.OutputStream{}

	// clamp loglevel to the range we're using
	logLevel := max(slog.Level(args.LogLevel), slog.LevelDebug)
	logLevel = min(logLevel, slog.LevelError)
	log.SetDefaultLogLevel(logLevel)

	if args.Cmd != "" {
		cmd, args := config.ParseCmdStr(args.Cmd)
		s := streams.NewCmdStream(context.Background(), "cmd", cmd, args...)
		inStreams = append(inStreams, s)
	}
	if args.OutputPath != "" {
		out, err := streams.NewOutputFile(args.OutputPath)
		if err != nil {
			log.Default().Error("invalid output path",
				slog.String("path", args.OutputPath),
				slog.String("err", err.Error()),
			)
			return nil, nil
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

	return
}

func runloop(inStreams []streams.InputStream, outStreams []streams.OutputStream) {

	a := streams.NewStreamAggregator(context.Background(), inStreams)

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
