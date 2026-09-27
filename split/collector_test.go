package split

import (
	"bytes"
	"math"
	"testing"

	"stegano2/bitio"
	"stegano2/mode"
	"stegano2/payload"
	"stegano2/rescode"
)

const (
	KEY_TEST        = "123"
	KEY_WRONG       = "321"
	NAME_TEST       = "data.bin"
	TEXT_DATA_TEST  = "скрытые данные для проверки"
	CNT_CELLS_FRAME = 1200
	CNT_BITS_EXTRA  = 777
	CNT_LIMIT_SMALL = 20000
)

// streamBits строит поток и раскладывает
// его на биты с добавкой лишних ячеек.
func streamBits(test *testing.T) ([]byte, []uint8) {
	stream, resCode := payload.Build(
		NAME_TEST,
		[]byte(TEXT_DATA_TEST),
		KEY_TEST,
		mode.MODES[0],
	)
	if resCode != rescode.OK {
		test.Fatalf("код %d", resCode)
	}
	cntBits := bitio.CntBits(len(stream))
	bits := make([]uint8, cntBits+CNT_BITS_EXTRA)
	for indBit := 0; indBit < cntBits; indBit++ {
		bits[indBit] = bitio.Get(stream, indBit)
	}
	return stream, bits
}

// feed подаёт биты накопителю кадрами.
func feed(taker *collector, bits []uint8) int {
	cntFrames := 0
	indFirst := 0
	for indFirst < len(bits) && taker.isActive() {
		indLast := min(
			indFirst+CNT_CELLS_FRAME, len(bits),
		)
		taker.take(bits[indFirst:indLast])
		indFirst = indLast
		cntFrames++
	}
	return cntFrames
}

func TestCollectorGathersStream(test *testing.T) {
	stream, bits := streamBits(test)
	taker := newCollector(KEY_TEST, math.MaxInt)
	cntFrames := feed(taker, bits)
	cntNeed := bitio.CntBits(len(stream))
	cntFramesRef := (cntNeed + CNT_CELLS_FRAME - 1) /
		CNT_CELLS_FRAME
	isRight := taker.isDone &&
		taker.resCode == rescode.OK &&
		bytes.Equal(taker.stream, stream) &&
		cntFrames == cntFramesRef
	if !isRight {
		test.Fatalf("код %d, кадров %d из %d",
			taker.resCode, cntFrames, cntFramesRef)
	}
}

func TestCollectorRejects(test *testing.T) {
	_, bits := streamBits(test)
	takerKey := newCollector(KEY_WRONG, math.MaxInt)
	takerSmall := newCollector(
		KEY_TEST, CNT_LIMIT_SMALL,
	)
	feed(takerKey, bits)
	feed(takerSmall, bits)
	isRight := !takerKey.isDone &&
		takerKey.resCode ==
			rescode.ERR_STREAM_HEADER &&
		!takerSmall.isDone &&
		takerSmall.resCode ==
			rescode.ERR_STREAM_SHORT
	if !isRight {
		test.Fatalf("коды %d %d",
			takerKey.resCode, takerSmall.resCode)
	}
}
