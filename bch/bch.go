//============================================//
// Пакет bch                                  //
//                                            //
// Назначение:                                //
//   помехоустойчивый код                     //
//   Боуза — Чоудхури — Хоквингема (127, 64): //
//   64 бита данных, 63 проверочных бита,     //
//   исправляет до 10 битовых ошибок в блоке. //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package bch

import (
	"encoding/binary"
	"math/bits"
)

// Encode вычисляет проверочные биты блока
// систематическим делением на порождающий
// полином.
//
// Параметры:
//   - data: 64 бита данных; старший бит —
//     коэффициент при старшей степени.
//
// Возвращает: 63 проверочных бита.
func Encode(data uint64) uint64 {
	register := uint64(0)
	feedback := GENERATOR & MASK_PARITY
	for indBit := CNT_BITS_DATA - 1; indBit >= 0; indBit-- {
		bitData := (data >> indBit) & 1
		bitTop := register >> (CNT_BITS_PARITY - 1)
		register = (register << 1) & MASK_PARITY
		if bitData^bitTop != 0 {
			register ^= feedback
		}
	}
	return register
}

// addTerms добавляет в синдромы вклад
// единичных битов одного слова.
//
// Параметры:
//   - syndromes: накапливаемые синдромы,
//     индексы 1..20.
//   - word: биты слова.
//   - shift: степень младшего бита слова.
func addTerms(
	syndromes *[CNT_SYNDROMES + 1]uint8,
	word uint64,
	shift int,
) {
	for rest := word; rest != 0; rest &= rest - 1 {
		degree := bits.TrailingZeros64(rest) + shift
		for indSyn := 1; indSyn <= CNT_SYNDROMES; indSyn++ {
			syndromes[indSyn] ^=
				powAlpha(indSyn * degree)
		}
	}
}

// findLocator строит полином локаторов ошибок
// алгоритмом Берлекэмпа — Месси.
//
// Параметры:
//   - syndromes: синдромы, индексы 1..20.
//
// Возвращает: коэффициенты полинома
// (индекс — степень) и его степень.
func findLocator(
	syndromes *[CNT_SYNDROMES + 1]uint8,
) ([SIZE_LOCATOR]uint8, int) {
	var sigma, prev [SIZE_LOCATOR]uint8
	sigma[0], prev[0] = 1, 1
	lenLfsr, shift := 0, 1
	discPrev := uint8(1)
	for indStep := 0; indStep < CNT_SYNDROMES; indStep++ {
		disc := syndromes[indStep+1]
		for indTap := 1; indTap <= lenLfsr; indTap++ {
			disc ^= mulField(
				sigma[indTap],
				syndromes[indStep+1-indTap],
			)
		}
		saved := sigma
		scale := divField(disc, discPrev)
		for indCoef := shift; indCoef < SIZE_LOCATOR; indCoef++ {
			sigma[indCoef] ^= mulField(
				scale,
				prev[indCoef-shift],
			)
		}
		shift++
		if disc != 0 && lenLfsr<<1 <= indStep {
			lenLfsr = indStep + 1 - lenLfsr
			prev = saved
			discPrev = disc
			shift = 1
		}
	}
	return sigma, lenLfsr
}

// evalLocator вычисляет значение полинома
// локаторов в точке по схеме Горнера.
//
// Параметры:
//   - sigma: коэффициенты полинома.
//   - degree: степень полинома, 0..21.
//   - point: точка, элемент поля.
//
// Возвращает: значение, элемент поля.
func evalLocator(
	sigma *[SIZE_LOCATOR]uint8,
	degree int,
	point uint8,
) uint8 {
	value := uint8(0)
	for indCoef := degree; indCoef >= 0; indCoef-- {
		value = mulField(value, point) ^
			sigma[indCoef]
	}
	return value
}

// maskData возвращает маску бита данных
// для позиции кодового слова.
//
// Параметры:
//   - indBit: степень позиции, 0..126.
//
// Возвращает: маску бита данных или 0,
// если позиция проверочная.
func maskData(indBit int) uint64 {
	mask := uint64(0)
	if indBit >= CNT_BITS_PARITY {
		mask = uint64(1) << (indBit - CNT_BITS_PARITY)
	}
	return mask
}

// Decode исправляет до 10 ошибок в блоке.
//
// Параметры:
//   - data: 64 принятых бита данных.
//   - parity: 63 принятых проверочных бита.
//
// Возвращает: исправленные данные (при
// неудаче — принятые без изменений), число
// исправленных битов блока и признак успеха.
func Decode(
	data uint64,
	parity uint64,
) (uint64, int, bool) {
	var syndromes [CNT_SYNDROMES + 1]uint8
	addTerms(&syndromes, parity&MASK_PARITY, 0)
	addTerms(&syndromes, data, CNT_BITS_PARITY)
	sigma, degree := findLocator(&syndromes)

	flipData := uint64(0)
	cntRoots := 0
	isFit := degree <= CNT_ERRORS_MAX
	for indBit := 0; isFit && indBit < CNT_BITS_CODE; indBit++ {
		point := powAlpha(SIZE_GROUP - indBit)
		if evalLocator(&sigma, degree, point) == 0 {
			cntRoots++
			flipData |= maskData(indBit)
		}
	}
	isOk := isFit && cntRoots == degree
	if !isOk {
		flipData, cntRoots = 0, 0
	}
	return data ^ flipData, cntRoots, isOk
}

// EncodeBlock кодирует 8 байтов данных
// в блок из 16 байтов: данные, затем
// 63 проверочных бита и нулевой бит.
//
// Параметры:
//   - src: ровно 8 байтов данных.
//
// Возвращает: блок из 16 байтов.
func EncodeBlock(src []byte) [SIZE_BLOCK]byte {
	var block [SIZE_BLOCK]byte
	data := binary.BigEndian.Uint64(src)
	copy(block[:SIZE_DATA], src)
	binary.BigEndian.PutUint64(
		block[SIZE_DATA:],
		Encode(data)<<1,
	)
	return block
}

// DecodeBlock восстанавливает 8 байтов данных
// из блока в 16 байтов.
//
// Параметры:
//   - block: ровно 16 байтов блока.
//
// Возвращает: 8 байтов данных, число
// исправленных битов и признак успеха.
func DecodeBlock(
	block []byte,
) ([SIZE_DATA]byte, int, bool) {
	var dst [SIZE_DATA]byte
	data := binary.BigEndian.Uint64(block)
	tail := binary.BigEndian.Uint64(
		block[SIZE_DATA:],
	)
	fixed, cntFixed, isOk := Decode(data, tail>>1)
	binary.BigEndian.PutUint64(dst[:], fixed)
	return dst, cntFixed, isOk
}
