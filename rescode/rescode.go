//============================================//
// Пакет rescode                              //
//                                            //
// Назначение:                                //
//   единый перечень числовых кодов           //
//   результата для всех пакетов программы.   //
//   Ноль — успех, остальное — ошибки.        //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package rescode

// Успех
const OK uint8 = 0

// Ошибки аргументов командной строки
const (
	ERR_ARG_COUNT   uint8 = 1
	ERR_ARG_COMMAND uint8 = 2
	ERR_ARG_FLAG    uint8 = 3
	ERR_ARG_MISSING uint8 = 4
	ERR_ARG_KEY     uint8 = 5
	ERR_ARG_MODE    uint8 = 6
	ERR_ARG_PATH    uint8 = 7
)

// Ошибки скрываемых данных
const (
	ERR_DATA_EMPTY uint8 = 10
	ERR_DATA_SIZE  uint8 = 11
	ERR_DATA_CRC   uint8 = 12
)

// Ошибки файлов
const (
	ERR_FILE_READ   uint8 = 20
	ERR_FILE_RENAME uint8 = 21
	ERR_FILE_WRITE  uint8 = 22
)

// Ошибки анализа видео
const (
	ERR_PROBE_PARSE uint8 = 30
	ERR_PROBE_RATE  uint8 = 31
	ERR_PROBE_RUN   uint8 = 32
	ERR_PROBE_SIZE  uint8 = 33
	ERR_PROBE_VIDEO uint8 = 34
)

// Ошибки процессов и каналов
const (
	ERR_PIPE_CREATE uint8 = 40
	ERR_PIPE_WRITE  uint8 = 41
	ERR_PROC_AUDIO  uint8 = 42
	ERR_PROC_DECODE uint8 = 43
	ERR_PROC_ENCODE uint8 = 44
	ERR_PROC_LOOKUP uint8 = 45
	ERR_PROC_START  uint8 = 46
	ERR_PROC_STOP   uint8 = 47
)

// Ошибки потока скрытых данных
const (
	ERR_STREAM_HEADER uint8 = 50
	ERR_STREAM_SHORT  uint8 = 51
	ERR_STREAM_WRAP   uint8 = 52
)

// Chain выполняет шаги по порядку: первая
// ошибка останавливает цепочку.
//
// Параметры:
//   - steps: шаги, возвращающие код
//     результата.
//
// Возвращает: OK либо код первой ошибки.
func Chain(steps ...func() uint8) uint8 {
	resCode := OK
	indStep := 0
	for resCode == OK && indStep < len(steps) {
		resCode = steps[indStep]()
		indStep++
	}
	return resCode
}
