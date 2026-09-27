package mode

// Идентификаторы режимов
const (
	ID_AGGRESSIVE   = 1
	ID_EXPERIMENTAL = 2
	ID_COMFORTABLE  = 3
)

// Имена режимов
const (
	NAME_AGGRESSIVE   = "aggressive"
	NAME_EXPERIMENTAL = "experimental"
	NAME_COMFORTABLE  = "comfortable"
)

// Доля узора в смеси с исходным кадром
const (
	MIX_AGGRESSIVE   = 0.5
	MIX_EXPERIMENTAL = 0.1
	MIX_COMFORTABLE  = 0.05
)

// Число копий данных в потоке
const (
	CNT_COPIES_AGGRESSIVE   = 25
	CNT_COPIES_EXPERIMENTAL = 15
	CNT_COPIES_COMFORTABLE  = 10
)

// Качество сжатия видео (CRF кодека H.264)
const (
	CRF_AGGRESSIVE   = "18"
	CRF_EXPERIMENTAL = "18"
	CRF_COMFORTABLE  = "16"
)
