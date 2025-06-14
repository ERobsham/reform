package streams

import (
	"bufio"
	"bytes"
	"testing"
)

const (
	ex1 = `Jun 12 08:24:46 hst-name0000 abc[34798]: <Debug> CoolClient received response: {
	URI = "state/update";
	response = "update request received";
} <line:000563 file:/src/common/CoolClient.m>
Jun 12 08:24:47 hst-name0000 abc[34798]: <Debug> NNG Socket connected <line:000308 file:/src/common/nng/nngWrapper.m>`

	ex2 = `Jun 12 08:21:28.12034 hst-name000A abc-go[119421]: utils/info.go:138: Initialized info : &{Type:33 Model:ABC-SDF-00 PartNumber:00-11-22-33 ID:001AAE1}
Jun 12 08:21:28.12896 hst-name000A abc-go[119421]: api/server.go:243: Registering endpoint modules`
)

func Test_readNextLine(t *testing.T) {
	type args struct {
		pipe *bufio.Reader
	}
	tests := []struct {
		name     string
		args     args
		wantLine string
		wantErr  bool
	}{
		{
			name:     "multiline ex 1",
			args:     args{bufio.NewReader(bytes.NewBufferString(ex1))},
			wantLine: "Jun 12 08:24:46 hst-name0000 abc[34798]: <Debug> CoolClient received response: { URI = \"state/update\"; response = \"update request received\"; } <line:000563 file:/src/common/CoolClient.m>",
			wantErr:  false,
		},
		{
			name:     "ex 2",
			args:     args{bufio.NewReader(bytes.NewBufferString(ex2))},
			wantLine: "Jun 12 08:21:28.12034 hst-name000A abc-go[119421]: utils/info.go:138: Initialized info : &{Type:33 Model:ABC-SDF-00 PartNumber:00-11-22-33 ID:001AAE1}",
			wantErr:  false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, err := readNextLine(tt.args.pipe)
			if (err != nil) != tt.wantErr {
				t.Errorf("readNextChunk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLine != tt.wantLine {
				t.Errorf("readNextChunk() gotLine vs wantLine:\n  %v\n  %v", gotLine, tt.wantLine)
			}
		})
	}
}
