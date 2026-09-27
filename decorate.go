package main

import (
	"fmt"

	"stegano2/report"
	"stegano2/rescode"
)

// isUsageError сообщает, вызвана ли ошибка
// неверной командной строкой.
//
// Параметры:
//   - resCode: код результата.
//
// Возвращает: true для ошибок ERR_ARG_*,
// кроме отсутствия команды.
func isUsageError(resCode uint8) bool {
	return resCode > rescode.ERR_ARG_COUNT &&
		resCode <= rescode.ERR_ARG_PATH
}

// decorate дополняет сообщение: для ошибки —
// её кодом, пояснением внешней программы и
// справкой, если ошибка в командной строке;
// после строки хода работы — переводом
// строки.
//
// Параметры:
//   - message: текст сообщения.
//   - result: отчёт о выполнении.
//
// Возвращает: текст для печати.
func decorate(
	message string,
	result report.Report,
) string {
	isFailure := result.Code != rescode.OK &&
		result.Code != rescode.ERR_ARG_COUNT
	if isFailure {
		message = fmt.Sprintf(
			FMT_FAILURE, result.Code, message,
		)
	}
	if isFailure && len(result.Detail) > 0 {
		message = fmt.Sprintf(
			FMT_DETAIL, message, result.Detail,
		)
	}
	if isUsageError(result.Code) {
		message = fmt.Sprintf(
			FMT_HINT, message, MSG_USAGE,
		)
	}
	if isProgressShown {
		message = fmt.Sprintf(
			FMT_DETAIL, VAL_EMPTY, message,
		)
	}
	return message
}
