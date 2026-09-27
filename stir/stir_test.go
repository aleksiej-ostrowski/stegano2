package stir

import (
	"bytes"
	"math/rand"
	"testing"
)

const (
	SEED_TEST      = 20260927
	SIZE_DATA_TEST = 10007
	KEY_TEST       = "123"
	KEY_WRONG      = "124"
	VALUE_TEST     = 117
	SEED_LEGACY    = 101
)

// randomData создаёт воспроизводимые данные.
func randomData() []byte {
	generator := rand.New(rand.NewSource(SEED_TEST))
	data := make([]byte, SIZE_DATA_TEST)
	generator.Read(data)
	return data
}

func TestRoundTrip(test *testing.T) {
	data := randomData()
	for _, isBitShuffle := range []bool{false, true} {
		mixed := Shuffle(data, KEY_TEST, isBitShuffle)
		restored := Unshuffle(
			mixed, KEY_TEST, isBitShuffle,
		)
		if bytes.Equal(mixed, data) ||
			!bytes.Equal(restored, data) {
			test.Fatalf("биты=%v", isBitShuffle)
		}
	}
}

func TestWrongKeyFails(test *testing.T) {
	data := randomData()
	mixed := Shuffle(data, KEY_TEST, true)
	restored := Unshuffle(mixed, KEY_WRONG, true)
	if bytes.Equal(restored, data) {
		test.Fatalf("чужой ключ подошёл")
	}
}

func TestShuffleKeepsBytes(test *testing.T) {
	data := randomData()
	mixed := Shuffle(data, KEY_TEST, false)
	var cntSrc, cntDst [1 << CNT_BITS_BYTE]int
	for indByte := range data {
		cntSrc[data[indByte]]++
		cntDst[mixed[indByte]]++
	}
	if cntSrc != cntDst {
		test.Fatalf("состав байтов изменился")
	}
}

func TestBitShuffleInverse(test *testing.T) {
	forward := rand.New(rand.NewSource(SEED_LEGACY))
	backward := rand.New(rand.NewSource(SEED_LEGACY))
	mixed := shuffleBits(VALUE_TEST, forward)
	restored := unshuffleBits(mixed, backward)
	if restored != VALUE_TEST {
		test.Fatalf("%d -> %d -> %d",
			VALUE_TEST, mixed, restored)
	}
}
