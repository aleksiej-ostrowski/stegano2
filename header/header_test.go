package header

import (
	"math/rand"
	"testing"

	"stegano2/rescode"
)

const (
	SEED_TEST       = 20260927
	KEY_TEST        = "123"
	KEY_WRONG       = "321"
	ID_MODE_TEST    = 2
	SIZE_TEST       = 292176
	CRC_TEST        = 0xDEADBEEF
	CNT_BITS_BYTE   = 8
	CNT_PERCENT     = 100
	CNT_NOISE_SHARE = 25
	SIZE_STREAM_REF = 1488
)

// sampleHeader возвращает образец заголовка.
func sampleHeader() Header {
	return Header{
		IdMode:      ID_MODE_TEST,
		SizeWrapped: SIZE_TEST,
		CrcData:     CRC_TEST,
	}
}

func TestRoundTrip(test *testing.T) {
	stream := Encode(sampleHeader(), KEY_TEST)
	header, resCode := Decode(stream, KEY_TEST)
	isSame := resCode == rescode.OK &&
		header == sampleHeader() &&
		len(stream) == SIZE_STREAM_REF
	if !isSame {
		test.Fatalf("код %d: %+v", resCode, header)
	}
}

func TestSurvivesNoise(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	stream := Encode(sampleHeader(), KEY_TEST)
	cntBits := len(stream) * CNT_BITS_BYTE
	cntFlips := cntBits * CNT_NOISE_SHARE / CNT_PERCENT
	order := generator.Perm(cntBits)
	for indFlip := 0; indFlip < cntFlips; indFlip++ {
		indBit := order[indFlip]
		stream[indBit/CNT_BITS_BYTE] ^=
			1 << (indBit % CNT_BITS_BYTE)
	}
	header, resCode := Decode(stream, KEY_TEST)
	if resCode != rescode.OK ||
		header != sampleHeader() {
		test.Fatalf("код %d: %+v", resCode, header)
	}
}

func TestWrongKey(test *testing.T) {
	stream := Encode(sampleHeader(), KEY_TEST)
	_, codeKey := Decode(stream, KEY_WRONG)
	_, codeSize := Decode(stream[1:], KEY_TEST)
	if codeKey != rescode.ERR_STREAM_HEADER ||
		codeSize != rescode.ERR_STREAM_HEADER {
		test.Fatalf("коды %d %d", codeKey, codeSize)
	}
}
