package payload

import (
	"bytes"
	"math/rand"
	"testing"

	"stegano2/header"
	"stegano2/mode"
	"stegano2/rescode"
)

const (
	SEED_TEST       = 20260927
	SIZE_DATA_TEST  = 3001
	NAME_TEST       = "данные.bin"
	KEY_TEST        = "123"
	KEY_WRONG       = "321"
	CNT_BITS_BYTE   = 8
	CNT_PERCENT     = 100
	CNT_NOISE_SHARE = 10
)

// randomData создаёт воспроизводимые данные.
func randomData(generator *rand.Rand) []byte {
	data := make([]byte, SIZE_DATA_TEST)
	generator.Read(data)
	return data
}

func TestRoundTripModes(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	data := randomData(generator)
	for _, params := range mode.MODES {
		stream, codeBuild := Build(
			NAME_TEST, data, KEY_TEST, params,
		)
		result, codeParse := Parse(stream, KEY_TEST)
		isSame := codeBuild == rescode.OK &&
			codeParse == rescode.OK &&
			bytes.Equal(result.Content.Data, data) &&
			result.Content.Name == NAME_TEST &&
			result.Params == params
		if !isSame {
			test.Fatalf("режим %s: %d %d",
				params.Name, codeBuild, codeParse)
		}
	}
}

func TestSurvivesNoise(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	data := randomData(generator)
	for _, params := range mode.MODES {
		stream, _ := Build(
			NAME_TEST, data, KEY_TEST, params,
		)
		cntBits := len(stream) * CNT_BITS_BYTE
		cntFlips := cntBits * CNT_NOISE_SHARE /
			CNT_PERCENT
		order := generator.Perm(cntBits)
		for indFlip := 0; indFlip < cntFlips; indFlip++ {
			indBit := order[indFlip]
			stream[indBit/CNT_BITS_BYTE] ^=
				1 << (indBit % CNT_BITS_BYTE)
		}
		result, resCode := Parse(stream, KEY_TEST)
		isSame := resCode == rescode.OK &&
			bytes.Equal(result.Content.Data, data) &&
			result.Stats.CntBitsFixed > 0
		if !isSame {
			test.Fatalf("режим %s: код %d",
				params.Name, resCode)
		}
	}
}

func TestRejects(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	data := randomData(generator)
	params := mode.MODES[0]
	stream, _ := Build(
		NAME_TEST, data, KEY_TEST, params,
	)
	_, codeEmpty := Build(
		NAME_TEST, nil, KEY_TEST, params,
	)
	_, codeKey := Parse(stream, KEY_WRONG)
	_, codeShort := Parse(
		stream[:len(stream)-1], KEY_TEST,
	)
	_, codeHead := Parse(
		stream[:header.SIZE_STREAM-1], KEY_TEST,
	)
	isRight := codeEmpty == rescode.ERR_DATA_EMPTY &&
		codeKey == rescode.ERR_STREAM_HEADER &&
		codeShort == rescode.ERR_STREAM_SHORT &&
		codeHead == rescode.ERR_STREAM_HEADER
	if !isRight {
		test.Fatalf("коды %d %d %d %d", codeEmpty,
			codeKey, codeShort, codeHead)
	}
}
