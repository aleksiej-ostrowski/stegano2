package merge

import (
	"stegano2/bitio"
	"stegano2/cell"
	"stegano2/ffpipe"
	"stegano2/probe"
	"stegano2/rescode"
)

// Plan — план построения видео.
// Grid — сетка ячеек кадра; CntFramesPass —
// кадров в одном проходе исходного видео;
// CntPasses — число проходов; CntFrames —
// всего кадров в итоговом видео.
type Plan struct {
	Grid          cell.Grid
	CntFramesPass int
	CntPasses     int
	CntFrames     int
}

// MakePlan строит план: исходное видео
// повторяется целое число раз так, чтобы
// вместить весь поток скрытых данных.
//
// Параметры:
//   - info: сведения об исходном видео.
//   - cntBytesStream: размер потока, > 0.
//
// Возвращает: план и код результата:
// OK либо ERR_PROBE_SIZE (кадр меньше
// ячейки или нет кадров).
func MakePlan(
	info probe.Info,
	cntBytesStream int,
) (Plan, uint8) {
	var plan Plan
	resCode := rescode.OK
	width, height := info.Width, info.Height
	if info.IsTurned {
		width, height = height, width
	}
	plan.Grid = cell.NewGrid(width&^1, height&^1)
	plan.CntFramesPass = info.CntFrames
	cntCells := plan.Grid.CntCells()
	if cntCells == 0 || plan.CntFramesPass <= 0 {
		resCode = rescode.ERR_PROBE_SIZE
	}
	if resCode == rescode.OK {
		cntBits := bitio.CntBits(cntBytesStream)
		cntFramesData := (cntBits + cntCells - 1) /
			cntCells
		plan.CntPasses = (cntFramesData +
			plan.CntFramesPass - 1) /
			plan.CntFramesPass
		plan.CntFrames = plan.CntPasses *
			plan.CntFramesPass
	}
	return plan, resCode
}

// cntSamplesAt считает звуковые выборки
// от начала видео до заданного кадра.
//
// Параметры:
//   - info: сведения об исходном видео.
//   - cntFrames: число кадров, >= 0.
//
// Возвращает: число выборок одного канала.
func cntSamplesAt(
	info probe.Info,
	cntFrames int,
) int64 {
	return int64(cntFrames) *
		int64(info.RateAudio) *
		int64(info.RateDen) /
		int64(info.RateNum)
}

// SizeAudioPass вычисляет объём звука
// одного прохода. Границы проходов считаются
// от начала видео, поэтому ошибка округления
// не накапливается.
//
// Параметры:
//   - info: сведения об исходном видео.
//   - plan: план построения видео.
//   - indPass: номер прохода, >= 0.
//
// Возвращает: число байтов выборок s16le.
func SizeAudioPass(
	info probe.Info,
	plan Plan,
	indPass int,
) int64 {
	cntFirst := cntSamplesAt(
		info, indPass*plan.CntFramesPass,
	)
	cntLast := cntSamplesAt(
		info, (indPass+1)*plan.CntFramesPass,
	)
	sizeSample := int64(info.CntChannels) *
		ffpipe.SIZE_SAMPLE
	return (cntLast - cntFirst) * sizeSample
}

// DustBit возвращает псевдослучайный бит для
// ячейки, не занятой данными. Значение
// зависит только от ключа и номера бита.
//
// Параметры:
//   - seed: начальное значение из ключа.
//   - indBit: номер бита потока, >= 0.
//
// Возвращает: бит, 0 или 1.
func DustBit(seed uint64, indBit int) uint8 {
	value := seed + uint64(indBit)*MIX_DUST_GAMMA
	value ^= value >> SHIFT_DUST_HIGH
	value *= MIX_DUST_FIRST
	value ^= value >> SHIFT_DUST_MID
	value *= MIX_DUST_SECOND
	value ^= value >> SHIFT_DUST_LOW
	return uint8(value & 1)
}
