package parser

import (
	"reflect"
	"testing"
)

func Test_parseSourceFileInfo(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name          string
		args          args
		wantSfInfo    SourceFileInfo
		wantRemainder string
		wantErr       bool
	}{
		{
			name:          "example 1",
			args:          args{"Main.swift:4 (init(objectPath:serviceName:)), debugging: Using DBus, objectPath: /com/example/cool_agent for CoolAgent"},
			wantSfInfo:    SourceFileInfo{Language: "Swift", Filename: "Main.swift", LineNumber: 4},
			wantRemainder: "(init(objectPath:serviceName:)), debugging: Using DBus, objectPath: /com/example/cool_agent for CoolAgent",
			wantErr:       false,
		},
		{
			name:          "example 2",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <line:000890 file:/src/common/apps/acme/main.c>"},
			wantSfInfo:    SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 2-1",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <line: 000890 file:/src/common/apps/acme/main.c>"},
			wantSfInfo:    SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 3",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <file:/src/common/apps/acme/main.c line:000890>"},
			wantSfInfo:    SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 3-1",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <file:/src/common/apps/acme/main.c line: 000890>"},
			wantSfInfo:    SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 3-1",
			args:          args{"[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError> <file:/src/common/apps/acme/main.c line: 000890>"},
			wantSfInfo:    SourceFileInfo{Language: "C", Filename: "/src/common/apps/acme/main.c", LineNumber: 890},
			wantRemainder: "[incMsg] msg_p: 50/05/001AAE1253B40053 name=\"Cool Device\" <func:acme_deliveryError>",
			wantErr:       false,
		},
		{
			name:          "example 4-1",
			args:          args{"<socket to some-service did close> <line:001174 file:src/srvMan.m>"},
			wantSfInfo:    SourceFileInfo{Language: "Objective-C", Filename: "src/srvMan.m", LineNumber: 1174},
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
