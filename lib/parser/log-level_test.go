package parser

import "testing"

func Test_parseLogLevel(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name          string
		args          args
		wantLogLevel  string
		wantRemainder string
		wantErr       bool
	}{
		{
			name:          "std Rust log level",
			args:          args{" [DEBUG] ... some message ..."},
			wantLogLevel:  "debug",
			wantRemainder: "... some message ...",
		},
		{
			name:          "unwrapped log level",
			args:          args{" crit ... some message ..."},
			wantLogLevel:  "crit",
			wantRemainder: "... some message ...",
		},
		{
			name:          "too short",
			args:          args{"nil"},
			wantRemainder: "nil",
			wantErr:       true,
		},
		{
			name:          "not a well-known prefix",
			args:          args{"not a log level on line"},
			wantRemainder: "not a log level on line",
			wantErr:       true,
		},
		{
			name:          "misc blob data",
			args:          args{"0ad0faf87d687757da5df8"},
			wantRemainder: "0ad0faf87d687757da5df8",
			wantErr:       true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLogLevel, gotRemainder, err := parseLogLevel(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLogLevel != tt.wantLogLevel {
				t.Errorf("parseLogLevel() gotLogLevel = %v, want %v", gotLogLevel, tt.wantLogLevel)
			}
			if gotRemainder != tt.wantRemainder {
				t.Errorf("parseLogLevel() gotRemainder = %v, want %v", gotRemainder, tt.wantRemainder)
			}
		})
	}
}
