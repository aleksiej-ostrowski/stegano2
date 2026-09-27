package split

import (
	"stegano2/bitio"
	"stegano2/header"
	"stegano2/payload"
	"stegano2/rescode"
)

// collector копит биты потока, приходящие
// кадр за кадром. Сначала набирается
// заголовок; по нему становится известен
// полный размер потока.
type collector struct {
	key       string
	cntLimit  int
	stream    []byte
	cntBits   int
	cntNeed   int
	isPayload bool
	isDone    bool
	resCode   uint8
}

// newCollector создаёт накопитель.
//
// Параметры:
//   - key: ключ перемешивания.
//   - cntLimit: наибольшее число битов,
//     которое вмещает видео.
//
// Возвращает: накопитель, ждущий заголовок.
func newCollector(
	key string,
	cntLimit int,
) *collector {
	return &collector{
		key:      key,
		cntLimit: cntLimit,
		stream:   make([]byte, header.SIZE_STREAM),
		cntNeed: bitio.CntBits(
			header.SIZE_STREAM,
		),
	}
}

// isActive сообщает, нужны ли ещё биты.
func (taker *collector) isActive() bool {
	return !taker.isDone &&
		taker.resCode == rescode.OK
}

// advance переходит к следующей стадии:
// после заголовка — к данным, после
// данных — к завершению.
func (taker *collector) advance() {
	if taker.isPayload {
		taker.isDone = true
	}
	if !taker.isPayload {
		head, params, resCode := payload.ReadHeader(
			taker.stream, taker.key,
		)
		sizeStream := payload.SizeStream(head, params)
		cntNeed := bitio.CntBits(sizeStream)
		taker.resCode = resCode
		isShort := resCode == rescode.OK &&
			cntNeed > taker.cntLimit
		if isShort {
			taker.resCode = rescode.ERR_STREAM_SHORT
		}
		if taker.resCode == rescode.OK {
			grown := make([]byte, sizeStream)
			copy(grown, taker.stream)
			taker.stream = grown
			taker.cntNeed = cntNeed
		}
		taker.isPayload = true
	}
}

// take принимает биты очередного кадра.
//
// Параметры:
//   - bits: биты ячеек кадра, 0 или 1.
func (taker *collector) take(bits []uint8) {
	indCell := 0
	for indCell < len(bits) && taker.isActive() {
		cntTake := min(
			len(bits)-indCell,
			taker.cntNeed-taker.cntBits,
		)
		for indTake := 0; indTake < cntTake; indTake++ {
			bitio.Put(
				taker.stream,
				taker.cntBits+indTake,
				bits[indCell+indTake],
			)
		}
		taker.cntBits += cntTake
		indCell += cntTake
		if taker.cntBits == taker.cntNeed {
			taker.advance()
		}
	}
}
