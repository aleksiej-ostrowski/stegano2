//============================================//
// Пакет cli                                  //
//                                            //
// Назначение:                                //
//   разбор командной строки. Результат —     //
//   одна из команд: merge.Args, split.Args   //
//   либо Help; вызывающий различает их       //
//   переключателем по типу.                  //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package cli

import (
	"flag"
	"io"
	"path/filepath"
	"unicode/utf8"

	"stegano2/merge"
	"stegano2/mode"
	"stegano2/rescode"
	"stegano2/split"
)

// Help — команда вывода справки.
type Help struct{}

// Command — одна из команд программы:
// merge.Args, split.Args либо Help.
type Command any

// newFlags создаёт набор ключей команды,
// который ничего не печатает сам.
func newFlags(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(
		name, flag.ContinueOnError,
	)
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}
	return flags
}

// addFlag объявляет строковый ключ.
//
// Параметры:
//   - flags: набор ключей команды.
//   - name: имя ключа.
//   - preset: значение по умолчанию.
//
// Возвращает: указатель на значение ключа.
func addFlag(
	flags *flag.FlagSet,
	name string,
	preset string,
) *string {
	return flags.String(name, preset, VAL_EMPTY)
}

// isSamePath сообщает, указывают ли два
// пути на один файл.
func isSamePath(left, right string) bool {
	absLeft, errLeft := filepath.Abs(left)
	absRight, errRight := filepath.Abs(right)
	return errLeft == nil && errRight == nil &&
		absLeft == absRight
}

// checkCommon проверяет разбор ключей,
// наличие значений и длину ключа.
//
// Параметры:
//   - flags: набор ключей команды.
//   - rest: аргументы после имени команды.
//   - key: указатель на значение ключа.
//   - values: обязательные значения.
//
// Возвращает: OK, ERR_ARG_FLAG,
// ERR_ARG_MISSING либо ERR_ARG_KEY.
func checkCommon(
	flags *flag.FlagSet,
	rest []string,
	key *string,
	values ...*string,
) uint8 {
	resCode := rescode.OK
	errParse := flags.Parse(rest)
	if errParse != nil || flags.NArg() > 0 {
		resCode = rescode.ERR_ARG_FLAG
	}
	for _, value := range append(values, key) {
		isMissing := resCode == rescode.OK &&
			len(*value) == 0
		if isMissing {
			resCode = rescode.ERR_ARG_MISSING
		}
	}
	isShort := utf8.RuneCountInString(*key) <
		LIM_KEY_LENGTH
	if resCode == rescode.OK && isShort {
		resCode = rescode.ERR_ARG_KEY
	}
	return resCode
}

// parseMerge разбирает ключи команды merge.
//
// Параметры:
//   - rest: аргументы после имени команды.
//
// Возвращает: merge.Args и код результата:
// OK либо ERR_ARG_*.
func parseMerge(rest []string) (Command, uint8) {
	var args merge.Args
	flags := newFlags(NAME_COMMAND_MERGE)
	key := addFlag(flags, NAME_FLAG_KEY, VAL_EMPTY)
	nameMode := addFlag(
		flags, NAME_FLAG_MODE, mode.NAME_AGGRESSIVE,
	)
	data := addFlag(flags, NAME_FLAG_DATA, VAL_EMPTY)
	original := addFlag(
		flags, NAME_FLAG_ORIGINAL, VAL_EMPTY,
	)
	result := addFlag(
		flags, NAME_FLAG_RESULT, VAL_EMPTY,
	)
	resCode := checkCommon(
		flags, rest, key, data, original, result,
	)
	params, isMode := mode.ByName(*nameMode)
	if resCode == rescode.OK && !isMode {
		resCode = rescode.ERR_ARG_MODE
	}
	isSame := isSamePath(*result, *original) ||
		isSamePath(*result, *data)
	if resCode == rescode.OK && isSame {
		resCode = rescode.ERR_ARG_PATH
	}
	args = merge.Args{
		Key:          *key,
		Params:       params,
		PathData:     *data,
		PathOriginal: *original,
		PathResult:   *result,
	}
	return args, resCode
}

// parseSplit разбирает ключи команды split.
//
// Параметры:
//   - rest: аргументы после имени команды.
//
// Возвращает: split.Args и код результата:
// OK либо ERR_ARG_*.
func parseSplit(rest []string) (Command, uint8) {
	flags := newFlags(NAME_COMMAND_SPLIT)
	key := addFlag(flags, NAME_FLAG_KEY, VAL_EMPTY)
	input := addFlag(
		flags, NAME_FLAG_INPUT, VAL_EMPTY,
	)
	output := addFlag(
		flags, NAME_FLAG_OUTPUT, VAL_EMPTY,
	)
	resCode := checkCommon(
		flags, rest, key, input, output,
	)
	isSame := isSamePath(*input, *output)
	if resCode == rescode.OK && isSame {
		resCode = rescode.ERR_ARG_PATH
	}
	args := split.Args{
		Key:        *key,
		PathInput:  *input,
		PathOutput: *output,
	}
	return args, resCode
}

// Parse разбирает командную строку.
//
// Параметры:
//   - args: аргументы без имени программы.
//
// Возвращает: команду и код результата:
// OK, ERR_ARG_COUNT (нет команды),
// ERR_ARG_COMMAND (неизвестная команда)
// либо иной ERR_ARG_*.
func Parse(args []string) (Command, uint8) {
	var command Command = Help{}
	resCode := rescode.ERR_ARG_COUNT
	name := VAL_EMPTY
	if len(args) > 0 {
		name = args[0]
		resCode = rescode.ERR_ARG_COMMAND
	}
	switch name {
	case NAME_COMMAND_MERGE:
		command, resCode = parseMerge(args[1:])
	case NAME_COMMAND_SPLIT:
		command, resCode = parseSplit(args[1:])
	case NAME_COMMAND_HELP:
		resCode = rescode.OK
	}
	return command, resCode
}
