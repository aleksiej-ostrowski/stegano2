//============================================//
// Пакет payload                              //
//                                            //
// Назначение:                                //
//   путь данных без видео. Сборка потока:    //
//     данные -> код БЧХ -> N копий ->        //
//     перемешивание по ключу -> заголовок.   //
//   Разбор потока выполняет обратные шаги    //
//   и сверяет контрольную сумму данных.      //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package payload

import (
	"hash/crc32"

	"stegano2/header"
	"stegano2/mode"
	"stegano2/repeat"
	"stegano2/rescode"
	"stegano2/stir"
	"stegano2/wrap"
)

// Result — итог разбора потока.
type Result struct {
	Content wrap.Content
	Params  mode.Params
	Stats   repeat.Stats
}

// Build собирает поток скрытых данных.
//
// Параметры:
//   - name: имя файла данных (без пути).
//   - data: скрываемые данные, не пустые.
//   - key: ключ перемешивания.
//   - params: параметры режима.
//
// Возвращает: поток (заголовок и данные)
// и код результата: OK, ERR_DATA_EMPTY
// либо ERR_DATA_SIZE.
func Build(
	name string,
	data []byte,
	key string,
	params mode.Params,
) ([]byte, uint8) {
	var stream, wrapped []byte
	resCode := rescode.OK
	sizeCopies := params.CntCopies *
		wrap.SizeWrapped(len(name), len(data))
	if sizeCopies > LIM_SIZE_STREAM {
		resCode = rescode.ERR_DATA_SIZE
	}
	if len(data) == 0 {
		resCode = rescode.ERR_DATA_EMPTY
	}
	if resCode == rescode.OK {
		wrapped, resCode = wrap.Wrap(name, data)
	}
	if resCode == rescode.OK {
		head := header.Header{
			IdMode:      params.Id,
			SizeWrapped: uint32(len(wrapped)),
			CrcData:     crc32.ChecksumIEEE(data),
		}
		copies := repeat.Repeat(
			wrapped, params.CntCopies,
		)
		stream = append(
			header.Encode(head, key),
			stir.Shuffle(
				copies, key, params.IsBitShuffle,
			)...,
		)
	}
	return stream, resCode
}

// ReadHeader разбирает заголовок потока.
//
// Параметры:
//   - stream: начало потока, не короче
//     header.SIZE_STREAM байтов.
//   - key: ключ перемешивания.
//
// Возвращает: заголовок, параметры режима и
// код результата: OK либо ERR_STREAM_HEADER.
func ReadHeader(
	stream []byte,
	key string,
) (header.Header, mode.Params, uint8) {
	var head header.Header
	var params mode.Params
	resCode := rescode.OK
	if len(stream) < header.SIZE_STREAM {
		resCode = rescode.ERR_STREAM_HEADER
	}
	if resCode == rescode.OK {
		head, resCode = header.Decode(
			stream[:header.SIZE_STREAM], key,
		)
	}
	if resCode == rescode.OK {
		found, isFound := mode.ById(head.IdMode)
		params = found
		if !isFound {
			resCode = rescode.ERR_STREAM_HEADER
		}
	}
	return head, params, resCode
}

// SizeStream вычисляет полный размер потока.
//
// Параметры:
//   - head: заголовок потока.
//   - params: параметры режима.
//
// Возвращает: размер потока в байтах.
func SizeStream(
	head header.Header,
	params mode.Params,
) int {
	return header.SIZE_STREAM +
		int(head.SizeWrapped)*params.CntCopies
}

// Parse восстанавливает данные из потока.
//
// Параметры:
//   - stream: полный поток с заголовком.
//   - key: ключ перемешивания.
//
// Возвращает: итог разбора и код результата:
// OK, ERR_STREAM_HEADER, ERR_STREAM_SHORT,
// ERR_STREAM_WRAP либо ERR_DATA_CRC.
func Parse(
	stream []byte,
	key string,
) (Result, uint8) {
	var result Result
	var wrapped []byte
	head, params, resCode := ReadHeader(stream, key)
	if resCode == rescode.OK &&
		len(stream) != SizeStream(head, params) {
		resCode = rescode.ERR_STREAM_SHORT
	}
	if resCode == rescode.OK {
		result.Params = params
		copies := stir.Unshuffle(
			stream[header.SIZE_STREAM:],
			key,
			params.IsBitShuffle,
		)
		wrapped, result.Stats = repeat.Restore(
			copies, params.CntCopies,
		)
		result.Content, resCode = wrap.Unwrap(wrapped)
	}
	if resCode == rescode.OK {
		crcData := crc32.ChecksumIEEE(
			result.Content.Data,
		)
		if crcData != head.CrcData {
			resCode = rescode.ERR_DATA_CRC
		}
	}
	return result, resCode
}
