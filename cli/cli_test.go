package cli

import (
	"testing"

	"stegano2/merge"
	"stegano2/mode"
	"stegano2/rescode"
	"stegano2/split"
)

const (
	KEY_TEST      = "123"
	KEY_SHORT     = "12"
	PATH_DATA     = "./data/dubrowskij.txt"
	PATH_ORIGINAL = "./data/original.mp4"
	PATH_RESULT   = "./data/result.mp4"
	PATH_OUTPUT   = "./data/output.txt"
	NAME_UNKNOWN  = "unknown"
	FLAG_KEY      = "--key"
	FLAG_MODE     = "--mode"
	FLAG_DATA     = "--data"
	FLAG_ORIGINAL = "--original"
	FLAG_RESULT   = "--result"
	FLAG_INPUT    = "--input"
	FLAG_OUTPUT   = "--output"
	FLAG_UNKNOWN  = "--unknown"
)

// lineMerge строит командную строку merge.
func lineMerge(key, nameMode, result string) []string {
	return []string{
		NAME_COMMAND_MERGE,
		FLAG_KEY, key,
		FLAG_MODE, nameMode,
		FLAG_DATA, PATH_DATA,
		FLAG_ORIGINAL, PATH_ORIGINAL,
		FLAG_RESULT, result,
	}
}

func TestParseMerge(test *testing.T) {
	command, resCode := Parse(lineMerge(
		KEY_TEST, mode.NAME_COMFORTABLE, PATH_RESULT,
	))
	args, isMerge := command.(merge.Args)
	isRight := resCode == rescode.OK && isMerge &&
		args.Key == KEY_TEST &&
		args.Params.Id == mode.ID_COMFORTABLE &&
		args.PathData == PATH_DATA &&
		args.PathOriginal == PATH_ORIGINAL &&
		args.PathResult == PATH_RESULT
	if !isRight {
		test.Fatalf("код %d: %+v", resCode, command)
	}
}

func TestParseMergeDefaultMode(test *testing.T) {
	command, resCode := Parse([]string{
		NAME_COMMAND_MERGE,
		FLAG_KEY, KEY_TEST,
		FLAG_DATA, PATH_DATA,
		FLAG_ORIGINAL, PATH_ORIGINAL,
		FLAG_RESULT, PATH_RESULT,
	})
	args, isMerge := command.(merge.Args)
	isRight := resCode == rescode.OK && isMerge &&
		args.Params.Id == mode.ID_AGGRESSIVE
	if !isRight {
		test.Fatalf("код %d: %+v", resCode, command)
	}
}

func TestParseSplit(test *testing.T) {
	command, resCode := Parse([]string{
		NAME_COMMAND_SPLIT,
		FLAG_KEY, KEY_TEST,
		FLAG_INPUT, PATH_RESULT,
		FLAG_OUTPUT, PATH_OUTPUT,
	})
	args, isSplit := command.(split.Args)
	isRight := resCode == rescode.OK && isSplit &&
		args.Key == KEY_TEST &&
		args.PathInput == PATH_RESULT &&
		args.PathOutput == PATH_OUTPUT
	if !isRight {
		test.Fatalf("код %d: %+v", resCode, command)
	}
}

func TestParseErrors(test *testing.T) {
	cases := []struct {
		line []string
		code uint8
	}{
		{nil, rescode.ERR_ARG_COUNT},
		{
			[]string{NAME_UNKNOWN},
			rescode.ERR_ARG_COMMAND,
		},
		{
			[]string{NAME_COMMAND_SPLIT, FLAG_UNKNOWN},
			rescode.ERR_ARG_FLAG,
		},
		{
			[]string{NAME_COMMAND_SPLIT, PATH_DATA},
			rescode.ERR_ARG_FLAG,
		},
		{
			[]string{
				NAME_COMMAND_SPLIT, FLAG_KEY, KEY_TEST,
			},
			rescode.ERR_ARG_MISSING,
		},
		{
			lineMerge(
				KEY_SHORT,
				mode.NAME_AGGRESSIVE,
				PATH_RESULT,
			),
			rescode.ERR_ARG_KEY,
		},
		{
			lineMerge(
				KEY_TEST, NAME_UNKNOWN, PATH_RESULT,
			),
			rescode.ERR_ARG_MODE,
		},
		{
			lineMerge(
				KEY_TEST,
				mode.NAME_AGGRESSIVE,
				PATH_ORIGINAL,
			),
			rescode.ERR_ARG_PATH,
		},
		{
			[]string{NAME_COMMAND_HELP},
			rescode.OK,
		},
	}
	for indCase, item := range cases {
		_, resCode := Parse(item.line)
		if resCode != item.code {
			test.Fatalf("случай %d: код %d, нужен %d",
				indCase, resCode, item.code)
		}
	}
}
