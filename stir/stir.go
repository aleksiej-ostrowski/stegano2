//============================================//
// Пакет stir                                 //
//                                            //
// Назначение:                                //
//   перемешивание байтов потока и битов      //
//   внутри каждого байта псевдослучайной     //
//   перестановкой, заданной ключом.          //
//   Перестановки совпадают с прежней         //
//   версией программы.                       //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package stir

import (
	"hash/fnv"
	"math/rand"
)

// bitOrder — перестановка битов байта.
type bitOrder [CNT_BITS_BYTE]uint8

// SeedOf превращает ключ в начальное
// значение генератора (хеш FNV-1a).
//
// Параметры:
//   - key: ключ перемешивания.
//
// Возвращает: начальное значение.
func SeedOf(key string) int64 {
	hash := fnv.New64a()
	hash.Write([]byte(key))
	return int64(hash.Sum64())
}

// buildOrder строит перестановку байтов.
//
// Параметры:
//   - cntBytes: длина потока, >= 0.
//   - generator: генератор, меняет состояние.
//
// Возвращает: перестановку номеров байтов.
func buildOrder(
	cntBytes int,
	generator *rand.Rand,
) []uint32 {
	order := make([]uint32, cntBytes)
	for indByte := range order {
		order[indByte] = uint32(indByte)
	}
	generator.Shuffle(
		cntBytes,
		func(indLeft, indRight int) {
			order[indLeft], order[indRight] =
				order[indRight], order[indLeft]
		},
	)
	return order
}

// nextBitOrder строит очередную перестановку
// битов байта.
//
// Параметры:
//   - generator: генератор, меняет состояние.
//
// Возвращает: перестановку номеров битов.
func nextBitOrder(generator *rand.Rand) bitOrder {
	var order bitOrder
	for indBit := range order {
		order[indBit] = uint8(indBit)
	}
	generator.Shuffle(
		CNT_BITS_BYTE,
		func(indLeft, indRight int) {
			order[indLeft], order[indRight] =
				order[indRight], order[indLeft]
		},
	)
	return order
}

// keepBits оставляет биты байта на местах и
// не расходует генератор.
func keepBits(value uint8, _ *rand.Rand) uint8 {
	return value
}

// shuffleBits переставляет биты байта: бит
// с номером order[ind] встаёт на место ind.
func shuffleBits(
	value uint8,
	generator *rand.Rand,
) uint8 {
	shuffled := uint8(0)
	for indBit, indFrom := range nextBitOrder(generator) {
		shuffled |= (value >> indFrom & 1) << indBit
	}
	return shuffled
}

// unshuffleBits возвращает биты байта на
// исходные места.
func unshuffleBits(
	value uint8,
	generator *rand.Rand,
) uint8 {
	restored := uint8(0)
	for indBit, indFrom := range nextBitOrder(generator) {
		restored |= (value >> indBit & 1) << indFrom
	}
	return restored
}

// pickMixer выбирает преобразование битов.
//
// Параметры:
//   - isBitShuffle: переставлять ли биты.
//   - mixer: преобразование при перестановке.
//
// Возвращает: выбранное преобразование.
func pickMixer(
	isBitShuffle bool,
	mixer func(uint8, *rand.Rand) uint8,
) func(uint8, *rand.Rand) uint8 {
	picked := keepBits
	if isBitShuffle {
		picked = mixer
	}
	return picked
}

// Shuffle перемешивает поток по ключу.
//
// Параметры:
//   - src: исходный поток, до 2^32 байтов.
//   - key: ключ перемешивания.
//   - isBitShuffle: переставлять ли биты
//     внутри каждого байта.
//
// Возвращает: перемешанный поток той же
// длины.
func Shuffle(
	src []byte,
	key string,
	isBitShuffle bool,
) []byte {
	generator := rand.New(rand.NewSource(SeedOf(key)))
	order := buildOrder(len(src), generator)
	mixer := pickMixer(isBitShuffle, shuffleBits)
	dst := make([]byte, len(src))
	for indDst, indSrc := range order {
		dst[indDst] = mixer(src[indSrc], generator)
	}
	return dst
}

// Unshuffle восстанавливает порядок потока.
//
// Параметры:
//   - src: перемешанный поток.
//   - key: ключ перемешивания.
//   - isBitShuffle: переставлялись ли биты
//     внутри каждого байта.
//
// Возвращает: поток в исходном порядке.
func Unshuffle(
	src []byte,
	key string,
	isBitShuffle bool,
) []byte {
	generator := rand.New(rand.NewSource(SeedOf(key)))
	order := buildOrder(len(src), generator)
	mixer := pickMixer(isBitShuffle, unshuffleBits)
	dst := make([]byte, len(src))
	for indSrc, indDst := range order {
		dst[indDst] = mixer(src[indSrc], generator)
	}
	return dst
}
