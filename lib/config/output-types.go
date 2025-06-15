package config

import (
	"encoding/json"
	"fmt"
)

//#< output_type

const (
	OutputTypeKey_Stdout = "stdout"
	OutputTypeKey_File   = "file"
)

const (
	OutputType_None OutputType = iota
	OutputType_Stdout
	OutputType_File
)

type OutputType uint8

func (o *OutputType) UnmarshalJSON(d []byte) error {
	var str string
	if err := json.Unmarshal(d, &str); err != nil {
		return err
	}

	switch str {
	case OutputTypeKey_Stdout:
		*o = OutputType_Stdout
	case OutputTypeKey_File:
		*o = OutputType_File
	default:
		*o = OutputType_None
		return fmt.Errorf("unknown OutputType")
	}

	return nil
}

//#> output_type

//#< output_type -- file

type OutputFileCfg struct {
	Path string
}

func ParseOutputFileCfg(cfg map[string]any) (OutputFileCfg, error) {
	path, ok := cfg["path"].(string)
	if !ok {
		return OutputFileCfg{}, OutputTypeParseError("missing required 'path' key")
	}
	return OutputFileCfg{Path: path}, nil
}

//#> output_type -- file

//#< output_type -- custom errors

const (
	ErrOutputCfgParseFailed OutputTypeParseError = "specific output_type config parse error"
)

type OutputTypeParseError string

func (e OutputTypeParseError) Error() string { return string(e) }
func (e OutputTypeParseError) Is(other error) bool {
	_, ok := other.(OutputTypeParseError)
	return ok
}

//#> output_type -- custom errors
