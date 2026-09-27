package bch

// Параметры кода БЧХ(127, 64)
const (
	CNT_BITS_CODE   = 127
	CNT_BITS_DATA   = 64
	CNT_BITS_PARITY = 63
	CNT_ERRORS_MAX  = 10
	CNT_SYNDROMES   = 20
)

// Параметры поля GF(2^7), полином x^7 + x^3 + 1
const (
	SIZE_FIELD     = 128
	SIZE_GROUP     = 127
	SIZE_LOCATOR   = 22
	POLY_FIELD     = 0x89
	MASK_FIELD_TOP = 0x80
)

// Размеры и маски блоков
const (
	SIZE_DATA   = 8
	SIZE_BLOCK  = 16
	MASK_PARITY = (uint64(1) << CNT_BITS_PARITY) - 1
)
