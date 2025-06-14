package parser

import (
	"reflect"
	"testing"
	"time"
)

func Test_ParseSystemTimeStamp(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name       string
		args       args
		wantRemain string
		wantErr    bool
	}{
		{
			name:       "std mac timestamp",
			args:       args{"Jan  2 03:04:05 sav-d011e5db9aee0000 ..."},
			wantRemain: "sav-d011e5db9aee0000 ...",
			wantErr:    false,
		},
		{
			name:       "std linux timestamp",
			args:       args{"Jan  2 03:04:05.66778 sav-d011e5db9aee0000 ..."},
			wantRemain: "sav-d011e5db9aee0000 ...",
			wantErr:    false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotRemain, err := ParseSystemTimeStamp(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeStamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, time.Time{}) {
				t.Errorf("parseTimeStamp() got zero value for time.Time")
			}
			if gotRemain != tt.wantRemain {
				t.Errorf("parseTimeStamp() gotRemain = %v, want %v", gotRemain, tt.wantRemain)
			}
		})
	}
}

func Test_pickMorePreciseTime(t *testing.T) {
	type args struct {
		timestamp time.Time
		other     time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "nano more precise than micro",
			args: args{
				time.Date(2025, 10, 10, 10, 10, 10, 123_123_123, time.UTC),
				time.Date(2025, 10, 10, 10, 10, 10, 123_123_000, time.UTC),
			},
			want: time.Date(2025, 10, 10, 10, 10, 10, 123_123_123, time.UTC),
		},
		{
			name: "micro more precise than milli",
			args: args{
				time.Date(2025, 10, 10, 10, 10, 10, 123_000_000, time.UTC),
				time.Date(2025, 10, 10, 10, 10, 10, 123_123_000, time.UTC),
			},
			want: time.Date(2025, 10, 10, 10, 10, 10, 123_123_000, time.UTC),
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickMorePreciseTime(tt.args.timestamp, tt.args.other); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pickMorePreciseTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseMsgPrefixTimeStamp(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name          string
		args          args
		wantTimestamp time.Time
		wantRemaining string
		wantErr       bool
	}{
		{
			name:          "example go log",
			args:          args{"12:24:46 cmd/server.go:57: Starting"},
			wantTimestamp: time.Date(0, 1, 1, 12, 24, 46, 0, time.UTC),
			wantRemaining: "cmd/server.go:57: Starting",
		},
		{
			name:          "example rust log",
			args:          args{"12:24:46.332 ::: cool_crate::some_module::some_type: Starting"},
			wantTimestamp: time.Date(0, 1, 1, 12, 24, 46, 332_000_000, time.UTC),
			wantRemaining: "::: cool_crate::some_module::some_type: Starting",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTimestamp, gotRemaining, err := parseMsgPrefixTimeStamp(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMsgPrefixTimeStamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTimestamp, tt.wantTimestamp) {
				t.Errorf("parseMsgPrefixTimeStamp() gotTimestamp = %v, want %v", gotTimestamp, tt.wantTimestamp)
			}
			if gotRemaining != tt.wantRemaining {
				t.Errorf("parseMsgPrefixTimeStamp() gotRemaining = %v, want %v", gotRemaining, tt.wantRemaining)
			}
		})
	}
}

func Test_parseMsgSuffixTimeStamp(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name          string
		args          args
		wantTimestamp time.Time
		wantRemaining string
		wantErr       bool
	}{
		{
			name:          "suffix timestamp - no wrapper (shortest)",
			args:          args{"<Critical> <request socket to some-service did close> <line:001174 file:src/ipc/srvMan.m> Aug 1 10:10:23:2025-10-01 UTC"},
			wantTimestamp: time.Date(2025, 10, 1, 10, 10, 23, 0, time.UTC),
			wantRemaining: "<Critical> <request socket to some-service did close> <line:001174 file:src/ipc/srvMan.m>",
		},
		{
			name:          "suffix timestamp - wrapper (shortest)",
			args:          args{"<Critical> <request socket to some-service did close> <line:001174 file:src/ipc/srvMan.m> <Aug 1 10:10:23:2025-10-01 UTC>"},
			wantTimestamp: time.Date(2025, 10, 1, 10, 10, 23, 0, time.UTC),
			wantRemaining: "<Critical> <request socket to some-service did close> <line:001174 file:src/ipc/srvMan.m>",
		},
		{
			name:          "suffix timestamp - no wrapper (std)",
			args:          args{"<Critical> <request socket to some-service did close> <line:001174 file:src/ipc/srvMan.m> Aug 10 10:10:23:2025-10-10 UTC"},
			wantTimestamp: time.Date(2025, 10, 10, 10, 10, 23, 0, time.UTC),
			wantRemaining: "<Critical> <request socket to some-service did close> <line:001174 file:src/ipc/srvMan.m>",
		},
		{
			name:          "suffix timestamp - wrapper (std)",
			args:          args{"<Critical> <request socket to some-service did close> <line:001174 file:src/ipc/srvMan.m> <Aug 10 10:10:23:2025-10-10 UTC>"},
			wantTimestamp: time.Date(2025, 10, 10, 10, 10, 23, 0, time.UTC),
			wantRemaining: "<Critical> <request socket to some-service did close> <line:001174 file:src/ipc/srvMan.m>",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTimestamp, gotRemaining, err := parseMsgSuffixTimeStamp(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMsgSuffixTimeStamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTimestamp, tt.wantTimestamp) {
				t.Errorf("parseMsgSuffixTimeStamp() gotTimestamp = %v, wantTimestamp %v", gotTimestamp, tt.wantTimestamp)
			}
			if gotRemaining != tt.wantRemaining {
				t.Errorf("parseMsgSuffixTimeStamp() gotRemaining = %v, wantRemaining %v", gotRemaining, tt.wantRemaining)
			}
		})
	}
}
