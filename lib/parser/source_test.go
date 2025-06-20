package parser

import (
	"reflect"
	"testing"

	"github.com/erobsham/reform/lib/types"
)

func Test_parseSourceFileInfo(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name          string
		args          args
		wantSfInfo    types.SourceFileInfo
		wantRemainder string
		wantErr       bool
	}{
		{
			name:          "example 1",
			args:          args{"Main.swift:4 (init(objectPath:serviceName:)), debugging: Using DBus, objectPath: /com/example/cool_agent for CoolAgent"},
			wantSfInfo:    types.SourceFileInfo{Language: "Swift", Filename: "Main.swift", LineNumber: 4},
			wantRemainder: "(init(objectPath:serviceName:)), debugging: Using DBus, objectPath: /com/example/cool_agent for CoolAgent",
			wantErr:       false,
		},
		{
			name:          "example 2",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <line:000890 file:/src/common/apps/acme/main.c>"},
			wantSfInfo:    types.SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 2-1",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <line: 000890 file:/src/common/apps/acme/main.c>"},
			wantSfInfo:    types.SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 3",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <file:/src/common/apps/acme/main.c line:000890>"},
			wantSfInfo:    types.SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 3-1",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <file:/src/common/apps/acme/main.c line: 000890>"},
			wantSfInfo:    types.SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 3-1",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <file:/src/common/apps/acme/main.c line: 000890>"},
			wantSfInfo:    types.SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 4-1",
			args:          args{"<socket to some-service did close> <line:001174 file:src/srvMan.m>"},
			wantSfInfo:    types.SourceFileInfo{Language: "Objective-C", Filename: "src/srvMan.m", LineNumber: 1174},
			wantRemainder: "<socket to some-service did close>",
			wantErr:       false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSfInfo, gotRemainder, err := parseSourceFileInfo(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSourceFileInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSfInfo, tt.wantSfInfo) {
				t.Errorf("parseSourceFileInfo() gotSfInfo got vs want:\n  %v\n  %v", gotSfInfo, tt.wantSfInfo)
			}
			if gotRemainder != tt.wantRemainder {
				t.Errorf("parseSourceFileInfo() gotRemainder got vs want:\n  %v\n  %v", gotRemainder, tt.wantRemainder)
			}
		})
	}
}

func Test_startIdxOfFilepath(t *testing.T) {
	type args struct {
		chunk string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "simple case",
			args: args{"source_file_name"},
			want: 0,
		},
		{
			name: "large path",
			args: args{"some/info <line:001459 file:/src/common/engine/ipc/cool_ipc"},
			want: 28,
		},
		{
			name: "relative path",
			args: args{"<socket to some-service did close> <line:001174 file:src/srvMan"},
			want: 53,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := startIdxOfFilepath(tt.args.chunk); got != tt.want {
				t.Errorf("startIdxOfFilepath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_consumeLineNumSuffix(t *testing.T) {
	type args struct {
		line   string
		endIdx int
	}
	tests := []struct {
		name        string
		args        args
		wantLineNum uint64
		wantIdx     int
	}{
		{
			name:        "line number suffix at EoL",
			args:        args{"src/file.c:10", 10},
			wantLineNum: 10,
			wantIdx:     13,
		},
		{
			name:        "line number suffix w/space",
			args:        args{"src/file.c:10 120 1203", 10},
			wantLineNum: 10,
			wantIdx:     13,
		},
		{
			name:        "line number suffix w/colon separator",
			args:        args{"src/file.c:10: 120 1203", 10},
			wantLineNum: 10,
			wantIdx:     13,
		},
		{
			name:        "line number suffix w/wrapper",
			args:        args{"src/file.c:10>120 1203", 10},
			wantLineNum: 10,
			wantIdx:     13,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLineNum, gotIdx := consumeLineNumSuffix(tt.args.line, tt.args.endIdx)
			if gotLineNum != tt.wantLineNum {
				t.Errorf("consumeLineNumSuffix() gotLineNum = %v, want %v", gotLineNum, tt.wantLineNum)
			}
			if gotIdx != tt.wantIdx {
				t.Errorf("consumeLineNumSuffix() gotIdx = %v, want %v", gotIdx, tt.wantIdx)
			}
		})
	}
}

func Test_consumeNextRustCrateName(t *testing.T) {
	type args struct {
		line     string
		startIdx int
	}
	tests := []struct {
		name      string
		args      args
		wantCrate string
		wantIdx   int
	}{
		{
			name:      "example 1",
			args:      args{"cool_crate::useful_module [WARN] something went wrong!", 0},
			wantCrate: "cool_crate::useful_module",
			wantIdx:   25,
		},
		{
			name:      "example 2",
			args:      args{"cool_crate::useful_module: [WARN] something went wrong!", 0},
			wantCrate: "cool_crate::useful_module",
			wantIdx:   25,
		},
		{
			name:      "example 3",
			args:      args{"cool_crate::useful_module] [WARN] something went wrong!", 0},
			wantCrate: "cool_crate::useful_module",
			wantIdx:   25,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCrate, gotIdx := consumeNextRustCrateName(tt.args.line, tt.args.startIdx)
			if gotCrate != tt.wantCrate {
				t.Errorf("consumeNextRustCrate() gotCrate = %v, want %v", gotCrate, tt.wantCrate)
			}
			if gotIdx != tt.wantIdx {
				t.Errorf("consumeNextRustCrate() gotIdx = %v, want %v", gotIdx, tt.wantIdx)
			}
		})
	}
}
