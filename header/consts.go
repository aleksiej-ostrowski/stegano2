package header

import "stegano2/bch"

// Подпись и версия формата потока
const (
	MAGIC   = "STG2"
	VERSION = 1
)

// Смещения полей заголовка
const (
	IDX_MAGIC    = 0
	IDX_VERSION  = 4
	IDX_MODE     = 5
	IDX_SIZE     = 8
	IDX_CRC_DATA = 12
	IDX_CRC_HEAD = 16
)

// Размеры
const (
	CNT_BLOCKS  = 3
	CNT_COPIES  = 31
	SIZE_RAW    = CNT_BLOCKS * bch.SIZE_DATA
	SIZE_CODED  = CNT_BLOCKS * bch.SIZE_BLOCK
	SIZE_STREAM = SIZE_CODED * CNT_COPIES
)
