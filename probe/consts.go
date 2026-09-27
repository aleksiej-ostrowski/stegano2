package probe

// Имена внешних программ
const NAME_FFPROBE = "ffprobe"

// Аргументы ffprobe
const (
	ARG_FORMAT      = "-print_format"
	ARG_LOG_LEVEL   = "-v"
	ARG_SHOW_FORMAT = "-show_format"
	ARG_SHOW_STREAM = "-show_streams"
	VAL_FORMAT_JSON = "json"
	VAL_LOG_ERROR   = "error"
)

// Значения полей отчёта ffprobe
const (
	VAL_TYPE_AUDIO  = "audio"
	VAL_TYPE_VIDEO  = "video"
	VAL_RATE_SPLIT  = "/"
	CNT_RATE_PARTS  = 2
	CNT_TURN_HALF   = 180
	CNT_TURN_QUART  = 90
	CNT_BITS_FLOAT  = 64
	LIM_ROUND_HALF  = 0.5
	SIZE_DETAIL_MAX = 2048
)
