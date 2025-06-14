package streams

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/erobsham/reform/lib/log"
	"github.com/erobsham/reform/lib/parser"
)

type OutputStream interface {
	Output(line parser.ParsedLine) error
	Close()
}

//#< File Output Stream

func NewOutputFile(file string) (OutputStream, error) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &OutputFile{file: f}, nil
}

type OutputFile struct {
	file *os.File
	lock sync.RWMutex
}

func (o *OutputFile) Output(line parser.ParsedLine) error {
	o.lock.RLock()
	defer o.lock.RUnlock()

	if o.file == nil {
		return ErrStreamClosed
	}

	encoder := json.NewEncoder(o.file)
	err := encoder.Encode(line)
	if err != nil {
		return err
	}

	return nil
}

func (o *OutputFile) Close() {
	o.lock.Lock()
	defer o.lock.Unlock()
	if o.file != nil {
		err := o.file.Sync()
		if err != nil {
			log.Default().
				Error("unable to sync outfile",
					slog.String("error", err.Error()),
				)
		}

		o.file.Close()
		o.file = nil
	}
}

//#> File Output Stream

//#< Stdout Stream

type StdoutStream struct {
	counter uint64
}

func (o *StdoutStream) Output(line parser.ParsedLine) error {
	fmt.Printf("%05d: %s \n", o.counter, line)
	o.counter += 1
	return nil
}

func (o StdoutStream) Close() {}

//#> Stdout Stream
