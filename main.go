//============================================//
// Программа stegano2                         //
//                                            //
// Назначение:                                //
//   стеганографические опыты с видео и       //
//   YouTube. Файл данных скрывается в кадрах //
//   видео узорами 8x8 пикселей и извлекается //
//   обратно даже после пережатия.            //
//                                            //
// Входные данные:                            //
//   merge: ключ, режим, файл данных,         //
//          исходное видео со звуком;         //
//   split: ключ, видео со скрытыми данными.  //
// Выходные данные:                           //
//   merge: видео со скрытыми данными         //
//          и звуковой дорожкой;              //
//   split: файл данных.                      //
//                                            //
// Весь вывод на экран — только в этом файле. //
//                                            //
// Автор: Aleksiej Ostrowski, 2023-2026       //
// https://aleksiej.com                       //
// Версия: 0.1.0                              //
//============================================//

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"stegano2/cli"
	"stegano2/merge"
	"stegano2/report"
	"stegano2/rescode"
	"stegano2/split"
)

// IS_TERMINAL — выводятся ли сообщения
// об ошибках на терминал.
var IS_TERMINAL = isTerminal(os.Stderr)

// isProgressShown — была ли напечатана
// строка хода работы.
var isProgressShown = false

// isTerminal сообщает, связан ли файл
// с терминалом.
//
// Параметры:
//   - file: открытый файл.
//
// Возвращает: true для терминала.
func isTerminal(file *os.File) bool {
	stat, errStat := file.Stat()
	return errStat == nil &&
		stat.Mode()&os.ModeCharDevice != 0
}

// showProgress печатает ход работы, если
// вывод идёт на терминал.
//
// Параметры:
//   - cntDone: число готовых кадров.
//   - cntTotal: общее число кадров.
func showProgress(cntDone, cntTotal int) {
	isStep := cntDone%CNT_PROGRESS_STEP == 0 ||
		cntDone == cntTotal
	if IS_TERMINAL && isStep {
		isProgressShown = true
		fmt.Fprintf(
			os.Stderr,
			FMT_PROGRESS,
			cntDone,
			cntTotal,
		)
	}
}

// wordAudio выбирает слово о звуке.
//
// Параметры:
//   - hasAudio: есть ли звуковая дорожка.
//
// Возвращает: слово для отчёта.
func wordAudio(hasAudio bool) string {
	word := MSG_AUDIO_NONE
	if hasAudio {
		word = MSG_AUDIO_KEPT
	}
	return word
}

// summaryMerge составляет отчёт об успешном
// сокрытии данных.
//
// Параметры:
//   - args: аргументы команды.
//   - result: отчёт о выполнении.
//
// Возвращает: текст отчёта.
func summaryMerge(
	args merge.Args,
	result report.Report,
) string {
	return fmt.Sprintf(
		FMT_MERGE,
		result.NameData,
		result.CntBytesData,
		result.NameMode,
		result.CntCopies,
		result.CntBytesStream,
		result.CntFrames,
		result.CntPasses,
		wordAudio(result.HasAudio),
		args.PathResult,
	)
}

// summarySplit составляет отчёт об успешном
// извлечении данных.
//
// Параметры:
//   - args: аргументы команды.
//   - result: отчёт о выполнении.
//
// Возвращает: текст отчёта.
func summarySplit(
	args split.Args,
	result report.Report,
) string {
	return fmt.Sprintf(
		FMT_SPLIT,
		result.NameData,
		result.CntBytesData,
		result.NameMode,
		result.CntCopies,
		wordAudio(result.HasAudio),
		result.CntBitsFixed,
		result.CntBitsBroken,
		result.CntBitsBch,
		result.CntBlocksFailed,
		args.PathOutput,
	)
}

// execute выполняет разобранную команду.
//
// Параметры:
//   - ctx: контекст отмены.
//   - command: merge.Args, split.Args
//     либо cli.Help.
//
// Возвращает: отчёт о выполнении и текст
// для печати при успехе.
func execute(
	ctx context.Context,
	command cli.Command,
) (report.Report, string) {
	var result report.Report
	summary := MSG_USAGE
	switch args := command.(type) {
	case merge.Args:
		result = merge.Run(ctx, args, showProgress)
		summary = summaryMerge(args, result)
	case split.Args:
		result = split.Run(ctx, args, showProgress)
		summary = summarySplit(args, result)
	}
	return result, summary
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	command, resCode := cli.Parse(os.Args[1:])
	result := report.Report{Code: resCode}
	summary := MSG_USAGE
	if resCode == rescode.OK {
		result, summary = execute(ctx, command)
	}
	stop()

	output := os.Stderr
	message := MSG_UNKNOWN
	switch result.Code {
	case rescode.OK:
		output = os.Stdout
		message = summary
	case rescode.ERR_ARG_COUNT:
		message = MSG_USAGE
	case rescode.ERR_ARG_COMMAND:
		message = MSG_ARG_COMMAND
	case rescode.ERR_ARG_FLAG:
		message = MSG_ARG_FLAG
	case rescode.ERR_ARG_MISSING:
		message = MSG_ARG_MISSING
	case rescode.ERR_ARG_KEY:
		message = MSG_ARG_KEY
	case rescode.ERR_ARG_MODE:
		message = MSG_ARG_MODE
	case rescode.ERR_ARG_PATH:
		message = MSG_ARG_PATH
	case rescode.ERR_DATA_EMPTY:
		message = MSG_DATA_EMPTY
	case rescode.ERR_DATA_SIZE:
		message = MSG_DATA_SIZE
	case rescode.ERR_DATA_CRC:
		message = MSG_DATA_CRC
	case rescode.ERR_FILE_READ:
		message = MSG_FILE_READ
	case rescode.ERR_FILE_RENAME:
		message = MSG_FILE_RENAME
	case rescode.ERR_FILE_WRITE:
		message = MSG_FILE_WRITE
	case rescode.ERR_PROBE_PARSE:
		message = MSG_PROBE_PARSE
	case rescode.ERR_PROBE_RATE:
		message = MSG_PROBE_RATE
	case rescode.ERR_PROBE_RUN:
		message = MSG_PROBE_RUN
	case rescode.ERR_PROBE_SIZE:
		message = MSG_PROBE_SIZE
	case rescode.ERR_PROBE_VIDEO:
		message = MSG_PROBE_VIDEO
	case rescode.ERR_PIPE_CREATE:
		message = MSG_PIPE_CREATE
	case rescode.ERR_PIPE_WRITE:
		message = MSG_PIPE_WRITE
	case rescode.ERR_PROC_AUDIO:
		message = MSG_PROC_AUDIO
	case rescode.ERR_PROC_DECODE:
		message = MSG_PROC_DECODE
	case rescode.ERR_PROC_ENCODE:
		message = MSG_PROC_ENCODE
	case rescode.ERR_PROC_LOOKUP:
		message = MSG_PROC_LOOKUP
	case rescode.ERR_PROC_START:
		message = MSG_PROC_START
	case rescode.ERR_PROC_STOP:
		message = MSG_PROC_STOP
	case rescode.ERR_STREAM_HEADER:
		message = MSG_STREAM_HEADER
	case rescode.ERR_STREAM_SHORT:
		message = MSG_STREAM_SHORT
	case rescode.ERR_STREAM_WRAP:
		message = MSG_STREAM_WRAP
	}
	message = decorate(message, result)
	fmt.Fprintln(output, message)
	os.Exit(int(result.Code))
}
