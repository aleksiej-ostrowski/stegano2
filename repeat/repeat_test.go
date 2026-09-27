package repeat

import (
	"bytes"
	"math/rand"
	"testing"
)

const (
	SEED_TEST        = 20260927
	SIZE_DATA_TEST   = 4096
	CNT_COPIES_TEST  = 25
	CNT_COPIES_SMALL = 3
	CNT_FLIPS_LIGHT  = 5
	CNT_FLIPS_HEAVY  = 12
	CNT_ACCEPT_TEST  = 17
	CNT_BITS_TEST    = SIZE_DATA_TEST * CNT_BITS_BYTE
)

// spoilBit инвертирует бит в заданном числе
// первых копий.
func spoilBit(
	stream []byte,
	indBit int,
	cntFlips int,
) {
	for indCopy := 0; indCopy < cntFlips; indCopy++ {
		indByte := indCopy*SIZE_DATA_TEST +
			indBit/CNT_BITS_BYTE
		stream[indByte] ^=
			1 << (indBit % CNT_BITS_BYTE)
	}
}

func TestRepeatRestoreClean(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	data := make([]byte, SIZE_DATA_TEST)
	generator.Read(data)
	stream := Repeat(data, CNT_COPIES_TEST)
	restored, stats := Restore(stream, CNT_COPIES_TEST)
	isClean := bytes.Equal(restored, data) &&
		stats.CntBitsFixed == 0 &&
		stats.CntBitsBroken == 0 &&
		len(stream) ==
			SIZE_DATA_TEST*CNT_COPIES_TEST
	if !isClean {
		test.Fatalf("статистика %+v", stats)
	}
}

func TestRestoreCountsVotes(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	data := make([]byte, SIZE_DATA_TEST)
	generator.Read(data)
	stream := Repeat(data, CNT_COPIES_TEST)
	order := generator.Perm(CNT_BITS_TEST)
	spoilBit(stream, order[0], CNT_FLIPS_LIGHT)
	spoilBit(stream, order[1], CNT_FLIPS_HEAVY)
	restored, stats := Restore(stream, CNT_COPIES_TEST)
	isRight := bytes.Equal(restored, data) &&
		stats.CntBitsFixed == 1 &&
		stats.CntBitsBroken == 1
	if !isRight {
		test.Fatalf("статистика %+v", stats)
	}
}

func TestAcceptThreshold(test *testing.T) {
	isRight := cntAcceptOf(CNT_COPIES_TEST) ==
		CNT_ACCEPT_TEST &&
		cntAcceptOf(CNT_COPIES_PAIR) ==
			CNT_COPIES_PAIR &&
		cntAcceptOf(1) == 1
	if !isRight {
		test.Fatalf("порог уверенности неверен")
	}
}

func TestRestoreMajority(test *testing.T) {
	stream := []byte{1, 1, 0}
	restored, stats := Restore(stream, CNT_COPIES_SMALL)
	isRight := len(restored) == 1 &&
		restored[0] == 1 &&
		stats.CntBitsFixed == 1
	if !isRight {
		test.Fatalf("%v %+v", restored, stats)
	}
}
