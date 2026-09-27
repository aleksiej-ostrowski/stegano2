//============================================//
// Пакет wrap                                 //
//                                            //
// Назначение:                                //
//   обёртка данных помехоустойчивым кодом    //
//   БЧХ(127, 64) с перемежением полубайтов.  //
//   Формат контейнера побайтно совместим     //
//   с прежним сценарием codilla.py:          //
//     блок 0   — длина имени (дважды),       //
//     блок 1   — длина данных (дважды),      //
//     блоки    — имя файла,                  //
//     блоки    — данные.                     //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package wrap

import (
	"encoding/binary"

	"stegano2/bch"
	"stegano2/rescode"
)

// Content — результат разбора контейнера.
type Content struct {
	Name            string
	Data            []byte
	CntBitsFixed    int
	CntBlocksFailed int
}

// cntBlocksOf считает блоки по 8 байтов,
// вмещающие заданное число байтов.
//
// Параметры:
//   - cntBytes: число байтов, >= 0.
//
// Возвращает: число блоков, >= 0.
func cntBlocksOf(cntBytes int) int {
	return (cntBytes + bch.SIZE_DATA - 1) /
		bch.SIZE_DATA
}

// SizeWrapped вычисляет размер контейнера.
//
// Параметры:
//   - cntName: длина имени в байтах, >= 0.
//   - cntData: длина данных в байтах, >= 0.
//
// Возвращает: размер контейнера в байтах.
func SizeWrapped(cntName, cntData int) int {
	cntBlocks := CNT_BLOCKS_HEAD +
		cntBlocksOf(cntName) +
		cntBlocksOf(cntData)
	return cntBlocks * bch.SIZE_BLOCK
}

// sizeField формирует поле длины: число
// записано дважды, младший байт первым.
//
// Параметры:
//   - cntBytes: длина, 0..LIM_SIZE_DATA.
//
// Возвращает: 8 байтов поля.
func sizeField(cntBytes int) []byte {
	field := make([]byte, bch.SIZE_DATA)
	for indCopy := 0; indCopy < CNT_SIZE_COPIES; indCopy++ {
		binary.LittleEndian.PutUint32(
			field[indCopy*SIZE_UINT32:],
			uint32(cntBytes),
		)
	}
	return field
}

// appendBlocks кодирует байты блоками по 8
// (последний дополняется нулями) и
// дописывает блоки по 16 байтов.
//
// Параметры:
//   - plain: накапливаемые блоки.
//   - src: кодируемые байты.
//
// Возвращает: plain с новыми блоками.
func appendBlocks(plain, src []byte) []byte {
	for indSrc := 0; indSrc < len(src); indSrc += bch.SIZE_DATA {
		var chunk [bch.SIZE_DATA]byte
		copy(chunk[:], src[indSrc:])
		block := bch.EncodeBlock(chunk[:])
		plain = append(plain, block[:]...)
	}
	return plain
}

// Wrap оборачивает данные кодом БЧХ и
// перемежает полубайты.
//
// Параметры:
//   - name: имя файла данных (без пути).
//   - data: данные, до LIM_SIZE_DATA байтов.
//
// Возвращает: контейнер и код результата:
// OK либо ERR_DATA_SIZE.
func Wrap(name string, data []byte) ([]byte, uint8) {
	var wrapped []byte
	resCode := rescode.OK
	isFit := len(data) <= LIM_SIZE_DATA &&
		len(name) <= LIM_SIZE_DATA
	if !isFit {
		resCode = rescode.ERR_DATA_SIZE
	}
	if resCode == rescode.OK {
		plain := make(
			[]byte,
			0,
			SizeWrapped(len(name), len(data)),
		)
		plain = appendBlocks(
			plain, sizeField(len(name)),
		)
		plain = appendBlocks(
			plain, sizeField(len(data)),
		)
		plain = appendBlocks(plain, []byte(name))
		plain = appendBlocks(plain, data)
		wrapped = joinNibbles(
			interleave(splitNibbles(plain)),
		)
	}
	return wrapped, resCode
}

// decodeBlocks исправляет и декодирует все
// блоки контейнера, копит статистику.
//
// Параметры:
//   - plain: блоки по 16 байтов.
//   - content: приёмник статистики.
//
// Возвращает: данные блоков, по 8 байтов.
func decodeBlocks(
	plain []byte,
	content *Content,
) []byte {
	cntBlocks := len(plain) / bch.SIZE_BLOCK
	decoded := make(
		[]byte,
		0,
		cntBlocks*bch.SIZE_DATA,
	)
	for indBlock := 0; indBlock < cntBlocks; indBlock++ {
		indSrc := indBlock * bch.SIZE_BLOCK
		chunk, cntFixed, isOk := bch.DecodeBlock(
			plain[indSrc : indSrc+bch.SIZE_BLOCK],
		)
		decoded = append(decoded, chunk[:]...)
		content.CntBitsFixed += cntFixed
		if !isOk {
			content.CntBlocksFailed++
		}
	}
	return decoded
}

// readSize читает одну копию поля длины.
//
// Параметры:
//   - decoded: данные блоков.
//   - indField: номер поля, 0..1.
//   - indCopy: номер копии, 0..1.
//
// Возвращает: значение длины.
func readSize(
	decoded []byte,
	indField int,
	indCopy int,
) int {
	offset := indField*bch.SIZE_DATA +
		indCopy*SIZE_UINT32
	return int(
		binary.LittleEndian.Uint32(decoded[offset:]),
	)
}

// pickSizes выбирает длины имени и данных,
// согласованные с размером контейнера,
// перебирая копии полей длины.
//
// Параметры:
//   - decoded: данные блоков.
//   - sizeWrapped: размер контейнера.
//
// Возвращает: длину имени, длину данных
// и признак того, что длины найдены.
func pickSizes(
	decoded []byte,
	sizeWrapped int,
) (int, int, bool) {
	cntName, cntData, isFound := 0, 0, false
	cntPairs := CNT_SIZE_COPIES * CNT_SIZE_COPIES
	for indPair := 0; !isFound && indPair < cntPairs; indPair++ {
		cntName = readSize(
			decoded,
			IDX_FIELD_NAME,
			indPair/CNT_SIZE_COPIES,
		)
		cntData = readSize(
			decoded,
			IDX_FIELD_DATA,
			indPair%CNT_SIZE_COPIES,
		)
		isFound = SizeWrapped(cntName, cntData) ==
			sizeWrapped
	}
	return cntName, cntData, isFound
}

// Unwrap восстанавливает данные контейнера.
//
// Параметры:
//   - wrapped: контейнер; длина кратна 16
//     и не меньше 32 байтов.
//
// Возвращает: содержимое со статистикой
// исправлений и код результата: OK либо
// ERR_STREAM_WRAP (размеры не согласованы).
func Unwrap(wrapped []byte) (Content, uint8) {
	var content Content
	resCode := rescode.OK
	sizeHead := CNT_BLOCKS_HEAD * bch.SIZE_BLOCK
	isSized := len(wrapped) >= sizeHead &&
		len(wrapped)%bch.SIZE_BLOCK == 0
	if !isSized {
		resCode = rescode.ERR_STREAM_WRAP
	}
	if resCode == rescode.OK {
		plain := joinNibbles(
			deinterleave(splitNibbles(wrapped)),
		)
		decoded := decodeBlocks(plain, &content)
		cntName, cntData, isFound := pickSizes(
			decoded, len(wrapped),
		)
		if isFound {
			indName := CNT_BLOCKS_HEAD * bch.SIZE_DATA
			indData := indName +
				cntBlocksOf(cntName)*bch.SIZE_DATA
			content.Name = string(
				decoded[indName : indName+cntName],
			)
			content.Data =
				decoded[indData : indData+cntData]
		} else {
			resCode = rescode.ERR_STREAM_WRAP
		}
	}
	return content, resCode
}
