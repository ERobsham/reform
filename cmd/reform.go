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
	flag.StringVar(&a.SeqServer, "seq", "", "specify `{hostname}:{port}[;{apikey}]` ex: `localhost:5341` | `localhost:5341;api-key-value` (default: none)")

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
	if args.SeqServer != "" {
		host, key := config.ParseSeqServer(args.SeqServer)
		s := streams.NewSeqStream(context.Background(), host, key)
		outStreams = append(outStreams, s)
	}

	if args.ConfigPath != "" {
		ins, outs := handleConfig(args.ConfigPath)
		inStreams = append(inStreams, ins...)
		outStreams = append(outStreams, outs...)
	}

	if len(outStreams) == 0 {
		out := &streams.StdoutStream{}
		outStreams = append(outStreams, out)
	}

	return
}

func handleConfig(cfgPath string) (inStreams []streams.InputStream, outStreams []streams.OutputStream) {
	inStreams = []streams.InputStream{}
	outStreams = []streams.OutputStream{}

	cfg, err := config.LoadConfigFrom(cfgPath)
	if err != nil {
		log.Default().
			Error("error loading config json",
				slog.String("error", err.Error()),
			)
		return nil, nil
	}

	for name, src := range cfg.Sources {
		s := streams.NewCmdStream(context.Background(), name, src.Cmd, src.Args...)
		inStreams = append(inStreams, s)
	}

	for name, out := range cfg.Outputs {
		switch out.OutputType {
		case config.OutputType_Stdout:
			o := &streams.StdoutStream{}
			outStreams = append(outStreams, o)
		case config.OutputType_File:
			cfg, err := config.ParseOutputFileCfg(out.Config)
			if err != nil {
				log.Default().
					Error("file output config parsing error",
						slog.String("name", name),
						slog.String("error", err.Error()),
					)
				continue
			}

			o, err := streams.NewOutputFile(cfg.Path)
			if err != nil {
				log.Default().
					Error("error creating output file",
						slog.String("name", name),
						slog.String("error", err.Error()),
					)
				continue
			}
			outStreams = append(outStreams, o)
		// `case config.OutputType_None:`
		default:
			log.Default().
				Error("invalid output type",
					slog.String("name", name),
				)
			continue
		}
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
