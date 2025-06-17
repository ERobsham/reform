package streams

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/erobsham/reform/lib/log"
	"github.com/erobsham/reform/lib/parser"
)

const (
	ErrStreamClosed StreamError = "stream closed"
)

type InputStream interface {
	Next() (string, error)
}

func NewCmdStream(ctx context.Context, streamName string, command string, args ...string) CmdStream {
	if args == nil {
		args = []string{}
	}

	cmd := exec.Command(command, args...)

	s := CmdStream{
		name:   streamName,
		ctx:    ctx,
		cmd:    cmd,
		output: make(chan string),
	}

	go s.runloop()

	return s
}

type CmdStream struct {
	name   string
	ctx    context.Context
	cmd    *exec.Cmd
	output chan string
}

func (s CmdStream) Next() (string, error) {
	val, ok := <-s.output

	if !ok {
		return val, ErrStreamClosed
	} else {
		return val, nil
	}
}

func (s CmdStream) runloop() {
	defer close(s.output)

	pipe, err := s.cmd.StdoutPipe()
	if err != nil {
		return
	}
	errPipe, _ := s.cmd.StderrPipe()

	reader := bufio.NewReader(pipe)
	errReader := bufio.NewReader(errPipe)

	err = s.cmd.Start()
	if err != nil {
		return
	}

outer:
	for {
		line, err := readNextLine(reader)
		if err != nil && !errors.Is(err, io.EOF) {
			errOut, _ := errReader.ReadBytes('\n')
			if len(errOut) > 0 {
				log.Default().
					Error("cmd err",
						slog.String("error", string(errOut)),
					)
			}
			break
		}

		select {
		case <-s.ctx.Done():
			break outer
		case s.output <- line:
			if errors.Is(err, io.EOF) {
				break outer
			} else {
				continue
			}
		}
	}
}

func readNextLine(pipe *bufio.Reader) (string, error) {
	var line string
	for {
		str, err := pipe.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return line + str, err
			}
			return line, err
		}

		str = strings.TrimSpace(str)
		if line == "" {
			line = str
		} else {
			line += str
		}

		if !isStartOfLineOrEmpty(pipe) {
			line += " "
			continue
		}

		return line, err
	}
}

func isStartOfLineOrEmpty(pipe *bufio.Reader) bool {
	data, err := pipe.Peek(parser.SysTimestamp_max_len + 1)
	if err != nil {
		return true
	}

	_, _, err = parser.ParseSystemTimeStamp(string(data))
	return err == nil
}

type StreamError string

func (s StreamError) Error() string { return string(s) }
