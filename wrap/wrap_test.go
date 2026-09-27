package wrap

import (
	"bytes"
	"math/rand"
	"os"
	"testing"

	"stegano2/rescode"
)

const (
	PATH_SAMPLE      = "testdata/sample.txt"
	PATH_SAMPLE_WRAP = "testdata/sample.txt.crp"
	NAME_SAMPLE      = "sample.txt"
	SEED_TEST        = 20260927
	CNT_BITS_BYTE    = 8
	CNT_FLIPS_TEST   = 600
	CNT_SIZES_TEST   = 40
	SIZE_TRUNCATED   = 17
)

// loadSample читает образец и его эталонную
// обёртку, созданную сценарием codilla.py.
func loadSample(test *testing.T) ([]byte, []byte) {
	sample, errSample := os.ReadFile(PATH_SAMPLE)
	wrapped, errWrap := os.ReadFile(PATH_SAMPLE_WRAP)
	if errSample != nil || errWrap != nil {
		test.Fatalf("нет образцов: %v %v",
			errSample, errWrap)
	}
	return sample, wrapped
}

func TestWrapMatchesReference(test *testing.T) {
	sample, reference := loadSample(test)
	wrapped, resCode := Wrap(NAME_SAMPLE, sample)
	if resCode != rescode.OK {
		test.Fatalf("код %d", resCode)
	}
	if !bytes.Equal(wrapped, reference) {
		test.Fatalf("обёртка отличается от эталона")
	}
}

func TestUnwrapReference(test *testing.T) {
	sample, reference := loadSample(test)
	content, resCode := Unwrap(reference)
	if resCode != rescode.OK {
		test.Fatalf("код %d", resCode)
	}
	isSame := bytes.Equal(content.Data, sample) &&
		content.Name == NAME_SAMPLE &&
		content.CntBitsFixed == 0 &&
		content.CntBlocksFailed == 0
	if !isSame {
		test.Fatalf("содержимое: %+v", content)
	}
}

func TestUnwrapFixesNoise(test *testing.T) {
	sample, reference := loadSample(test)
	generator := rand.New(rand.NewSource(SEED_TEST))
	cntBits := len(reference) * CNT_BITS_BYTE
	order := generator.Perm(cntBits)
	for indFlip := 0; indFlip < CNT_FLIPS_TEST; indFlip++ {
		indBit := order[indFlip]
		reference[indBit/CNT_BITS_BYTE] ^=
			1 << (indBit % CNT_BITS_BYTE)
	}
	content, resCode := Unwrap(reference)
	if resCode != rescode.OK {
		test.Fatalf("код %d", resCode)
	}
	if !bytes.Equal(content.Data, sample) {
		test.Fatalf(
			"данные не восстановлены: %d сбоев",
			content.CntBlocksFailed,
		)
	}
	if content.CntBitsFixed == 0 {
		test.Fatalf("исправления не учтены")
	}
}

func TestRoundTripSizes(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	for cntData := 0; cntData < CNT_SIZES_TEST; cntData++ {
		data := make([]byte, cntData)
		generator.Read(data)
		name := NAME_SAMPLE[:cntData%len(NAME_SAMPLE)]
		wrapped, codeWrap := Wrap(name, data)
		content, codeUnwrap := Unwrap(wrapped)
		isSame := codeWrap == rescode.OK &&
			codeUnwrap == rescode.OK &&
			len(wrapped) ==
				SizeWrapped(len(name), cntData) &&
			bytes.Equal(content.Data, data) &&
			content.Name == name
		if !isSame {
			test.Fatalf("длина %d: %d %d",
				cntData, codeWrap, codeUnwrap)
		}
	}
}

func TestUnwrapRejectsBadSize(test *testing.T) {
	_, reference := loadSample(test)
	_, codeShort := Unwrap(
		reference[:SIZE_TRUNCATED],
	)
	_, codeCut := Unwrap(
		reference[:len(reference)>>1],
	)
	if codeShort != rescode.ERR_STREAM_WRAP ||
		codeCut != rescode.ERR_STREAM_WRAP {
		test.Fatalf("коды %d %d",
			codeShort, codeCut)
	}
}
