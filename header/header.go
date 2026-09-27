//============================================//
// Пакет header                               //
//                                            //
// Назначение:                                //
//   заголовок потока скрытых данных.         //
//   Хранит режим, размер контейнера и        //
//   контрольную сумму данных, поэтому при    //
//   извлечении не нужны ни режим, ни размер. //
//   Защита не зависит от режима: код БЧХ,    //
//   31 копия, перемешивание по ключу.        //
//                                            //
// Формат (24 байта):                         //
//    0..3   подпись "STG2"                   //
//    4      версия формата                   //
//    5      идентификатор режима             //
//    6..7   резерв                           //
//    8..11  размер контейнера, байты         //
//   12..15  CRC-32 исходных данных           //
//   16..19  CRC-32 байтов 0..15              //
//   20..23  резерв                           //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package header

import (
	"encoding/binary"
	"hash/crc32"

	"stegano2/bch"
	"stegano2/repeat"
	"stegano2/rescode"
	"stegano2/stir"
)

// Header — содержимое заголовка.
type Header struct {
	IdMode      uint8
	SizeWrapped uint32
	CrcData     uint32
}

// Encode формирует защищённый заголовок.
//
// Параметры:
//   - header: содержимое заголовка.
//   - key: ключ перемешивания.
//
// Возвращает: SIZE_STREAM байтов потока.
func Encode(header Header, key string) []byte {
	var raw [SIZE_RAW]byte
	copy(raw[IDX_MAGIC:], MAGIC)
	raw[IDX_VERSION] = VERSION
	raw[IDX_MODE] = header.IdMode
	binary.LittleEndian.PutUint32(
		raw[IDX_SIZE:], header.SizeWrapped,
	)
	binary.LittleEndian.PutUint32(
		raw[IDX_CRC_DATA:], header.CrcData,
	)
	binary.LittleEndian.PutUint32(
		raw[IDX_CRC_HEAD:],
		crc32.ChecksumIEEE(raw[:IDX_CRC_HEAD]),
	)
	coded := make([]byte, 0, SIZE_CODED)
	for indBlock := 0; indBlock < CNT_BLOCKS; indBlock++ {
		indRaw := indBlock * bch.SIZE_DATA
		block := bch.EncodeBlock(raw[indRaw:])
		coded = append(coded, block[:]...)
	}
	return stir.Shuffle(
		repeat.Repeat(coded, CNT_COPIES),
		key,
		true,
	)
}

// Decode восстанавливает заголовок и
// проверяет подпись, версию и CRC.
//
// Параметры:
//   - stream: SIZE_STREAM байтов потока.
//   - key: ключ перемешивания.
//
// Возвращает: заголовок и код результата:
// OK либо ERR_STREAM_HEADER (неверный ключ
// или в видео нет скрытых данных).
func Decode(
	stream []byte,
	key string,
) (Header, uint8) {
	var header Header
	var raw [SIZE_RAW]byte
	resCode := rescode.OK
	if len(stream) != SIZE_STREAM {
		resCode = rescode.ERR_STREAM_HEADER
	}
	if resCode == rescode.OK {
		coded, _ := repeat.Restore(
			stir.Unshuffle(stream, key, true),
			CNT_COPIES,
		)
		for indBlock := 0; indBlock < CNT_BLOCKS; indBlock++ {
			indCoded := indBlock * bch.SIZE_BLOCK
			chunk, _, _ := bch.DecodeBlock(
				coded[indCoded:],
			)
			copy(raw[indBlock*bch.SIZE_DATA:], chunk[:])
		}
		crcHead := binary.LittleEndian.Uint32(
			raw[IDX_CRC_HEAD:],
		)
		isValid := string(raw[:IDX_VERSION]) == MAGIC &&
			raw[IDX_VERSION] == VERSION &&
			crcHead == crc32.ChecksumIEEE(
				raw[:IDX_CRC_HEAD],
			)
		if !isValid {
			resCode = rescode.ERR_STREAM_HEADER
		}
	}
	if resCode == rescode.OK {
		header.IdMode = raw[IDX_MODE]
		header.SizeWrapped =
			binary.LittleEndian.Uint32(raw[IDX_SIZE:])
		header.CrcData =
			binary.LittleEndian.Uint32(
				raw[IDX_CRC_DATA:],
			)
	}
	return header, resCode
}
