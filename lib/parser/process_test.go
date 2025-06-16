package parser

import (
	"reflect"
	"testing"

	"github.com/erobsham/reform/lib/types"
)

func Test_parseProcessInfo(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name          string
		args          args
		wantProc      types.ProcessInfo
		wantRemainder string
		wantErr       bool
	}{
		{
			name:          "std proc w/PID",
			args:          args{"Process Name[112233]: ... message ..."},
			wantProc:      types.ProcessInfo{Name: "Process Name", PID: 112233},
			wantRemainder: "... message ...",
			wantErr:       false,
		},
		{
			name:          "std proc w/PID and TID",
			args:          args{"process[112][334]: ... message ..."},
			wantProc:      types.ProcessInfo{Name: "process", PID: 112, TID: 334},
			wantRemainder: "... message ...",
			wantErr:       false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProc, gotRemainder, err := parseProcessInfo(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseProcessInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotProc, tt.wantProc) {
				t.Errorf("parseProcessInfo() gotProc = %v, want %v", gotProc, tt.wantProc)
			}
			if gotRemainder != tt.wantRemainder {
				t.Errorf("parseProcessInfo() gotRemainder = %v, want %v", gotRemainder, tt.wantRemainder)
			}
		})
	}
}
