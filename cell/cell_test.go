package cell

import (
	"math/rand"
	"testing"
)

const (
	SEED_TEST         = 20260927
	SIZE_WIDTH_TEST   = 70
	SIZE_HEIGHT_TEST  = 36
	CNT_COLS_REF      = 8
	CNT_ROWS_REF      = 4
	SIZE_FRAME_REF    = 3780
	MIX_STRONG        = 0.5
	MIX_WEAK          = 0.05
	LEVEL_BACK_STEP   = 3
	LEVEL_BACK_BASE   = 40
	LEVEL_NOISE_RANGE = 5
	LEVEL_ON_REF      = 137
	LEVEL_OFF_REF     = 28
	LEVEL_CHROMA_REF  = 84
)

// makeFrame создаёт кадр с плавным фоном.
func makeFrame(grid Grid) []byte {
	pix := make([]byte, grid.SizeFrame())
	for indPix := range pix {
		rowPix := indPix / grid.Width
		pix[indPix] = uint8(
			LEVEL_BACK_BASE +
				rowPix/SIZE_CELL*LEVEL_BACK_STEP,
		)
	}
	return pix
}

// randomBits создаёт случайные биты ячеек.
func randomBits(
	grid Grid,
	generator *rand.Rand,
) []uint8 {
	bits := make([]uint8, grid.CntCells())
	for indCell := range bits {
		bits[indCell] = uint8(generator.Intn(CNT_PATTERNS))
	}
	return bits
}

// cntErrorsOf считает расхождения битов.
func cntErrorsOf(left, right []uint8) int {
	cntErrors := 0
	for indCell := range left {
		cntErrors += int(left[indCell] ^ right[indCell])
	}
	return cntErrors
}

func TestGrid(test *testing.T) {
	grid := NewGrid(SIZE_WIDTH_TEST, SIZE_HEIGHT_TEST)
	isRight := grid.CntCols == CNT_COLS_REF &&
		grid.CntRows == CNT_ROWS_REF &&
		grid.CntCells() ==
			CNT_COLS_REF*CNT_ROWS_REF &&
		grid.SizeFrame() == SIZE_FRAME_REF
	if !isRight {
		test.Fatalf("сетка %+v", grid)
	}
}

func TestPalette(test *testing.T) {
	palette := NewPalette(MIX_STRONG)
	isRight :=
		palette.LumaOn[LEVEL_BACK_BASE] ==
			LEVEL_ON_REF &&
			palette.LumaOff[LEVEL_BACK_BASE] ==
				LEVEL_OFF_REF &&
			palette.Chroma[LEVEL_BACK_BASE] ==
				LEVEL_CHROMA_REF &&
			palette.Chroma[LEVEL_CHROMA_MID] ==
				LEVEL_CHROMA_MID
	if !isRight {
		test.Fatalf("таблицы смешивания неверны")
	}
}

func TestDrawRecognize(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	grid := NewGrid(SIZE_WIDTH_TEST, SIZE_HEIGHT_TEST)
	for _, mix := range []float64{
		MIX_STRONG, MIX_WEAK,
	} {
		palette := NewPalette(mix)
		bits := randomBits(grid, generator)
		pix := makeFrame(grid)
		DrawFrame(
			pix, grid, &palette,
			func(indCell int) uint8 {
				return bits[indCell]
			},
		)
		found := RecognizeFrame(
			pix[:grid.SizeLuma()], grid,
		)
		if cntErrorsOf(bits, found) != 0 {
			test.Fatalf("доля узора %v", mix)
		}
	}
}

func TestRecognizeNoisy(test *testing.T) {
	generator := rand.New(rand.NewSource(SEED_TEST))
	grid := NewGrid(SIZE_WIDTH_TEST, SIZE_HEIGHT_TEST)
	palette := NewPalette(MIX_STRONG)
	bits := randomBits(grid, generator)
	pix := makeFrame(grid)
	DrawFrame(
		pix, grid, &palette,
		func(indCell int) uint8 {
			return bits[indCell]
		},
	)
	luma := pix[:grid.SizeLuma()]
	for indPix := range luma {
		luma[indPix] += uint8(
			generator.Intn(LEVEL_NOISE_RANGE),
		)
	}
	found := RecognizeFrame(luma, grid)
	if cntErrorsOf(bits, found) != 0 {
		test.Fatalf("шум сломал распознавание")
	}
}
