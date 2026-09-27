package cell

// Palette — таблицы смешивания уровней
// кадра с узором при заданной доле узора.
type Palette struct {
	LumaOff [CNT_LEVELS]uint8
	LumaOn  [CNT_LEVELS]uint8
	Chroma  [CNT_LEVELS]uint8
}

// NewPalette строит таблицы смешивания.
//
// Параметры:
//   - mix: доля узора в смеси, 0..1.
//
// Возвращает: таблицы для яркости тёмных и
// светлых пикселей узора и для цветности.
func NewPalette(mix float64) Palette {
	var palette Palette
	for level := range palette.Chroma {
		rest := float64(level) * (1 - mix)
		lumaOff := rest + LEVEL_LUMA_MIN*mix
		lumaOn := rest + LEVEL_LUMA_MAX*mix
		chroma := rest + LEVEL_CHROMA_MID*mix +
			LIM_ROUND_HALF
		palette.LumaOff[level] = uint8(lumaOff)
		palette.LumaOn[level] = uint8(lumaOn)
		palette.Chroma[level] = uint8(chroma)
	}
	return palette
}

// pickPatterns выбирает узоры ячеек одной
// строки сетки. Неполные ячейки на краях
// кадра данных не несут и получают узор
// по шахматному правилу.
//
// Параметры:
//   - patterns: приёмник узоров, по одному
//     на каждую ячейку строки.
//   - grid: сетка ячеек.
//   - rowCell: строка сетки, >= 0.
//   - bitOf: бит ячейки по её номеру.
func pickPatterns(
	patterns []func(int, int) bool,
	grid Grid,
	rowCell int,
	bitOf func(indCell int) uint8,
) {
	for colCell := range patterns {
		bit := uint8(rowCell+colCell) & 1
		isFull := rowCell < grid.CntRows &&
			colCell < grid.CntCols
		if isFull {
			bit = bitOf(rowCell*grid.CntCols + colCell)
		}
		patterns[colCell] = PATTERNS[bit]
	}
}

// DrawFrame накладывает узоры ячеек на кадр
// yuv420p: яркость смешивается с узором,
// цветность — с нейтральным серым.
//
// Параметры:
//   - pix: кадр, grid.SizeFrame() байтов;
//     изменяется на месте.
//   - grid: сетка ячеек.
//   - palette: таблицы смешивания.
//   - bitOf: бит ячейки по её номеру,
//     0..grid.CntCells()-1.
func DrawFrame(
	pix []byte,
	grid Grid,
	palette *Palette,
	bitOf func(indCell int) uint8,
) {
	cntColsAll := (grid.Width + SIZE_CELL - 1) /
		SIZE_CELL
	patterns := make(
		[]func(int, int) bool,
		cntColsAll,
	)
	for rowPix := 0; rowPix < grid.Height; rowPix++ {
		if rowPix%SIZE_CELL == 0 {
			pickPatterns(
				patterns,
				grid,
				rowPix/SIZE_CELL,
				bitOf,
			)
		}
		indRow := rowPix * grid.Width
		row := pix[indRow : indRow+grid.Width]
		for colPix, level := range row {
			isOn := patterns[colPix/SIZE_CELL](
				rowPix, colPix,
			)
			row[colPix] = palette.LumaOff[level]
			if isOn {
				row[colPix] = palette.LumaOn[level]
			}
		}
	}
	chroma := pix[grid.SizeLuma():grid.SizeFrame()]
	for indPix, level := range chroma {
		chroma[indPix] = palette.Chroma[level]
	}
}
