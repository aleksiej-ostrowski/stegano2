package bch

import (
	"math/rand"
	"testing"
)

// Эталон: порождающий полином galois.BCH(127, 64)
const (
	POLY_GENERATOR_REF = uint64(1)<<63 | 1<<61 |
		1<<56 | 1<<55 | 1<<53 | 1<<51 | 1<<49 |
		1<<48 | 1<<47 | 1<<40 | 1<<38 | 1<<36 |
		1<<35 | 1<<33 | 1<<32 | 1<<31 | 1<<30 |
		1<<26 | 1<<25 | 1<<24 | 1<<23 | 1<<22 |
		1<<21 | 1<<19 | 1<<18 | 1<<15 | 1<<5 |
		1<<2 | 1
	SEED_TEST       = 20260927
	CNT_ROUNDS_TEST = 2000
	CNT_ERRORS_OVER = 11
	CNT_DISTANCE    = 21
)

// flipBits инвертирует cntFlips различных
// случайных битов кодового слова.
func flipBits(
	data uint64,
	parity uint64,
	cntFlips int,
	generator *rand.Rand,
) (uint64, uint64) {
	order := generator.Perm(CNT_BITS_CODE)
	for indFlip := 0; indFlip < cntFlips; indFlip++ {
		indBit := order[indFlip]
		data ^= maskData(indBit)
		if indBit < CNT_BITS_PARITY {
			parity ^= uint64(1) << indBit
		}
	}
	return data, parity
}

func TestGeneratorMatchesReference(test *testing.T) {
	if GENERATOR != POLY_GENERATOR_REF {
		test.Fatalf(
			"полином %#x, эталон %#x",
			GENERATOR,
			uint64(POLY_GENERATOR_REF),
		)
	}
}

func TestDecodeFixesUpToLimit(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	for indRound := 0; indRound < CNT_ROUNDS_TEST; indRound++ {
		data := generator.Uint64()
		parity := Encode(data)
		cntFlips := indRound % (CNT_ERRORS_MAX + 1)
		badData, badParity := flipBits(
			data, parity, cntFlips, generator,
		)
		fixed, cntFixed, isOk := Decode(
			badData, badParity,
		)
		if !isOk || fixed != data ||
			cntFixed != cntFlips {
			test.Fatalf(
				"раунд %d: ошибок %d, "+
					"исправлено %d, успех %v",
				indRound, cntFlips, cntFixed, isOk,
			)
		}
	}
}

func TestDecodeNeverMiscorrectsSilently(
	test *testing.T,
) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	for indRound := 0; indRound < CNT_ROUNDS_TEST; indRound++ {
		data := generator.Uint64()
		parity := Encode(data)
		badData, badParity := flipBits(
			data, parity, CNT_ERRORS_OVER, generator,
		)
		fixed, cntFixed, isOk := Decode(
			badData, badParity,
		)
		isNear := cntFixed+CNT_ERRORS_OVER >=
			CNT_DISTANCE
		if isOk && !isNear {
			test.Fatalf(
				"раунд %d: ложное исправление %d",
				indRound, cntFixed,
			)
		}
		if !isOk && fixed != badData {
			test.Fatalf(
				"раунд %d: данные изменены",
				indRound,
			)
		}
	}
}

func TestBlockRoundTrip(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	for indRound := 0; indRound < CNT_ROUNDS_TEST; indRound++ {
		var src [SIZE_DATA]byte
		generator.Read(src[:])
		block := EncodeBlock(src[:])
		if block[SIZE_BLOCK-1]&1 != 0 {
			test.Fatalf("дополняющий бит не нулевой")
		}
		indByte := generator.Intn(SIZE_BLOCK)
		block[indByte] ^= MASK_FIELD_TOP
		dst, cntFixed, isOk := DecodeBlock(block[:])
		if !isOk || dst != src || cntFixed != 1 {
			test.Fatalf(
				"раунд %d: блок не восстановлен",
				indRound,
			)
		}
	}
}
