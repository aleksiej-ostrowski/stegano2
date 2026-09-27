package cell

// Размеры
const (
	SIZE_CELL        = 8
	SIZE_CELL_PIXELS = SIZE_CELL * SIZE_CELL
)

// Количества
const (
	CNT_LEVELS        = 256
	CNT_PATTERNS      = 2
	CNT_PLANES_CHROMA = 2
	CNT_SHIFT_CHROMA  = 1
	CNT_SHIFT_WORD    = 6
	CNT_WORDS_SET     = 4
	MASK_WORD_INDEX   = 63
)

// Уровни яркости и цветности кадра yuv420p
// (ограниченный диапазон)
const (
	LEVEL_CHROMA_MID = 128
	LEVEL_LUMA_MAX   = 235
	LEVEL_LUMA_MIN   = 16
)

// Пороги
const LIM_ROUND_HALF = 0.5
