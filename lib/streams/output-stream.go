package streams

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/erobsham/reform/lib/log"
	"github.com/erobsham/reform/lib/types"
)

type OutputStream interface {
	Output(line types.ParsedLine) error
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

func (o *OutputFile) Output(line types.ParsedLine) error {
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

func (o *StdoutStream) Output(line types.ParsedLine) error {
	fmt.Printf("%05d: %s \n", o.counter, line)
	o.counter += 1
	return nil
}

func (o StdoutStream) Close() {}

//#> Stdout Stream

//#< Seq Server Stream

type SeqStream struct {
	ctx    context.Context
	client *http.Client

	host   string
	apiKey string

	logChan chan types.ParsedLine

	wakeChan chan struct{}
	sendChan chan struct{}
}

func NewSeqStream(ctx context.Context, host string, apiKey string) OutputStream {
	client := http.Client{
		Transport: &http.Transport{
			IdleConnTimeout:       time.Second * 10,
			ResponseHeaderTimeout: time.Millisecond * 250,
			ExpectContinueTimeout: time.Millisecond * 250,
			TLSHandshakeTimeout:   time.Millisecond * 250,
			TLSClientConfig: &tls.Config{
				// allow talking HTTPS via self-signed certs
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Millisecond * 500,
	}

	s := &SeqStream{
		ctx:    ctx,
		client: &client,

		host:   host,
		apiKey: apiKey,

		logChan:  make(chan types.ParsedLine, 64),
		sendChan: make(chan struct{}, 1),
		wakeChan: make(chan struct{}, 1),
	}
	// preload chan so we can immediately take it when
	// we get out first message to log.
	s.sendChan <- struct{}{}

	go s.runloop()

	return s
}

func (s *SeqStream) Output(line types.ParsedLine) error {

	// block on adding the line to logChan
	select {
	case <-s.ctx.Done():
		return ErrStreamClosed
	case s.logChan <- line:
	}

	// don't block on signaling the wake up
	select {
	case <-s.ctx.Done():
		return ErrStreamClosed
	case s.wakeChan <- struct{}{}:
		return nil
	default:
		return nil
	}
}

func (s *SeqStream) Close() {

	// block for any pending sends to finish
	<-s.sendChan
	if len(s.logChan) > 0 {
		s.sendChan <- struct{}{}
		s.gatherAndSendLogs()
	}
	close(s.logChan)
	close(s.sendChan)
}

func (s *SeqStream) runloop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.wakeChan:
		}

		s.gatherAndSendLogs()
	}
}

func (s *SeqStream) gatherAndSendLogs() {
	lines := []types.ParsedLine{}
gather:
	for {
		select {
		case line := <-s.logChan:
			lines = append(lines, line)
		default:
			break gather
		}
	}

send:
	for {
		select {
		case <-s.sendChan:
			break send
		case line := <-s.logChan:
			lines = append(lines, line)
		}
	}

	err := s.sendLogs(lines)
	if err != nil {
		log.Default().
			Error("error sending logs to Seq",
				slog.String("error", err.Error()),
			)
	}
}

func (s *SeqStream) sendLogs(lines []types.ParsedLine) error {
	// signal we're done sending
	defer func() { s.sendChan <- struct{}{} }()

	if len(lines) == 0 {
		return nil
	}

	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	for _, line := range lines {
		encoder.Encode(line)
	}

	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.url(), buf)
	if err != nil {
		return err
	}

	// From the '[CLEF](https://clef-json.org/)' standard:
	// `ContentType: application/vnd.serilog.clef`
	// `X-Seq-ApiKey: {api key}`

	req.Header.Set("ContentType", "application/vnd.serilog.clef")
	if s.apiKey != "" {
		req.Header.Set("X-Seq-ApiKey", s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response status: %v body:%s", resp.StatusCode, string(body))
	}

	log.Default().Info("sent logs",
		slog.Int("numSent", len(lines)),
	)

	return nil
}

func (s *SeqStream) url() string {
	return "http://" + s.host + "/ingest/clef"
}

//#>
