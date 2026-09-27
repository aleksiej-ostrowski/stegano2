package wrap

import "math"

// Параметры перемежения полубайтов
const (
	CNT_BITS_NIBBLE  = 4
	CNT_NIBBLES_BYTE = 2
	CNT_STEP_NIBBLE  = 1980
	MASK_NIBBLE      = 0x0F
)

// Параметры заголовочных блоков
const (
	CNT_BLOCKS_HEAD = 2
	CNT_SIZE_COPIES = 2
	IDX_FIELD_DATA  = 1
	IDX_FIELD_NAME  = 0
	SIZE_UINT32     = 4
)

// Пороги
const LIM_SIZE_DATA = math.MaxUint32
