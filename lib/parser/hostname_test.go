package parser

import "testing"

func Test_parseHostName(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name          string
		args          args
		wantHostname  string
		wantRemainder string
		wantErr       bool
	}{
		{
			name:          "standard case",
			args:          args{"host-123abc SomeProcess[12394]: ..."},
			wantHostname:  "host-123abc",
			wantRemainder: "SomeProcess[12394]: ...",
			wantErr:       false,
		},
		{
			name:          "MacOS duplicates case",
			args:          args{"--- last message repeated 32 times ---"},
			wantHostname:  "",
			wantRemainder: "--- last message repeated 32 times ---",
			wantErr:       false,
		},
		{
			name:    "too short",
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHostname, gotRemainder, err := parseHostName(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHostName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHostname != tt.wantHostname {
				t.Errorf("parseHostName() gotHostname = %v, want %v", gotHostname, tt.wantHostname)
			}
			if gotRemainder != tt.wantRemainder {
				t.Errorf("parseHostName() gotRemainder = %v, want %v", gotRemainder, tt.wantRemainder)
			}
		})
	}
}
