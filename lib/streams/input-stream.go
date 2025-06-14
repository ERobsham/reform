package streams

import (
	"bufio"
	"context"
	"os/exec"
	"strings"

	"github.com/erobsham/reform/lib/parser"
)

const (
	ErrStreamClosed StreamError = "stream closed"
)

type InputStream interface {
	Next() (string, error)
}

func NewCmdStream(ctx context.Context, streamName string, cmdArgs []string) CmdStream {
	cmdName := cmdArgs[0]
	args := []string{}

	if len(cmdArgs) > 0 {
		args = cmdArgs[1:]
	}

	cmd := exec.Command(cmdName, args...)

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
	reader := bufio.NewReader(pipe)

	err = s.cmd.Start()
	if err != nil {
		return
	}

outer:
	for {
		line, err := readNextLine(reader)
		if err != nil {
			break
		}

		select {
		case <-s.ctx.Done():
			break outer
		case s.output <- line:
			continue
		}
	}
}

func readNextLine(pipe *bufio.Reader) (string, error) {
	var line string
	for {
		str, err := pipe.ReadString('\n')
		if err != nil {
			return line, err
		}

		str = strings.TrimSpace(str)
		if line == "" {
			line = str
		} else {
			line = line + " " + str
		}

		if isAnotherLineExpected(line) && !isStartOfLineOrEmpty(pipe) {
			continue
		}

		return line, err
	}
}

func isAnotherLineExpected(current string) bool {
	len := len(current)
	if len == 0 {
		return false
	}

	lastChar := current[len-1]
	possiblyJSON := isPossibleJSON(lastChar) || isPossibleJSONLike(lastChar)

	return possiblyJSON
}

func isStartOfLineOrEmpty(pipe *bufio.Reader) bool {
	data, err := pipe.Peek(parser.SysTimestamp_max_len + 1)
	if err != nil {
		return true
	}

	_, _, err = parser.ParseSystemTimeStamp(string(data))
	return err == nil
}

func isPossibleJSON(b byte) bool {
	lineEndings := map[byte]struct{}{
		'{': {}, // opening object
		'}': {}, // closing object

		'[': {}, // opening array
		']': {}, // closing array

		',': {}, // mid-object field separator

		'"': {}, // closing string field

		'0': {}, // closing number field
		'1': {}, // closing number field
		'2': {}, // closing number field
		'3': {}, // closing number field
		'4': {}, // closing number field
		'5': {}, // closing number field
		'6': {}, // closing number field
		'7': {}, // closing number field
		'8': {}, // closing number field
		'9': {}, // closing number field

		'e': {}, // closing bool field (tru'e' | fals'e')
		'l': {}, // closing null field (nul'l')
	}

	_, exists := lineEndings[b]

	return exists
}
func isPossibleJSONLike(b byte) bool {
	lineEndings := map[byte]struct{}{
		// Objective-C dictionary log like:
		// 	`{\n "key" = ( \n "value",\n "val2"\n );\n}`
		';': {},
		'=': {},
		'(': {},
		')': {},
	}

	_, exists := lineEndings[b]

	return exists
}

type StreamError string

func (s StreamError) Error() string { return string(s) }
