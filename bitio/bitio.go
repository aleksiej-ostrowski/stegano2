//============================================//
// Пакет bitio                                //
//                                            //
// Назначение:                                //
//   побитовый доступ к байтовому потоку.     //
//   Биты байта нумеруются от младшего.       //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package bitio

// CntBits считает биты потока.
//
// Параметры:
//   - cntBytes: длина потока в байтах, >= 0.
//
// Возвращает: число битов.
func CntBits(cntBytes int) int {
	return cntBytes << CNT_SHIFT_BYTE
}

// CntBytes считает байты, вмещающие биты.
//
// Параметры:
//   - cntBits: число битов, >= 0.
//
// Возвращает: число байтов.
func CntBytes(cntBits int) int {
	return (cntBits + MASK_BIT_INDEX) >>
		CNT_SHIFT_BYTE
}

// Get читает бит потока.
//
// Параметры:
//   - stream: поток байтов.
//   - indBit: номер бита,
//     0..CntBits(len(stream))-1.
//
// Возвращает: значение бита, 0 или 1.
func Get(stream []byte, indBit int) uint8 {
	value := stream[indBit>>CNT_SHIFT_BYTE]
	return value >> (indBit & MASK_BIT_INDEX) & 1
}

// Put добавляет единичный бит в поток
// (нулевой бит поток не меняет).
//
// Параметры:
//   - stream: поток байтов, изменяется.
//   - indBit: номер бита,
//     0..CntBits(len(stream))-1.
//   - bit: значение бита, 0 или 1.
func Put(stream []byte, indBit int, bit uint8) {
	stream[indBit>>CNT_SHIFT_BYTE] |=
		bit << (indBit & MASK_BIT_INDEX)
}
