package ffpipe

// Имена внешних программ
const (
	NAME_FFMPEG  = "ffmpeg"
	NAME_FFPROBE = "ffprobe"
)

// Ключи командной строки ffmpeg
const (
	ARG_BITRATE_AUDIO   = "-b:a"
	ARG_CHANNELS        = "-ac"
	ARG_CODEC_AUDIO     = "-c:a"
	ARG_CODEC_VIDEO     = "-c:v"
	ARG_COLOR_PRIMARIES = "-color_primaries"
	ARG_COLOR_RANGE     = "-color_range"
	ARG_COLOR_SPACE     = "-colorspace"
	ARG_COLOR_TRANSFER  = "-color_trc"
	ARG_CRF             = "-crf"
	ARG_FILTER_AUDIO    = "-af"
	ARG_FILTER_VIDEO    = "-vf"
	ARG_FORMAT          = "-f"
	ARG_FPS_MODE        = "-fps_mode"
	ARG_INPUT           = "-i"
	ARG_LOG_LEVEL       = "-v"
	ARG_MAP             = "-map"
	ARG_MOV_FLAGS       = "-movflags"
	ARG_NO_AUDIO        = "-an"
	ARG_NO_DATA         = "-dn"
	ARG_NO_ROTATE       = "-noautorotate"
	ARG_NO_STDIN        = "-nostdin"
	ARG_NO_SUBS         = "-sn"
	ARG_NO_VIDEO        = "-vn"
	ARG_OVERWRITE       = "-y"
	ARG_PIX_FMT         = "-pix_fmt"
	ARG_PRESET          = "-preset"
	ARG_RATE            = "-r"
	ARG_RATE_AUDIO      = "-ar"
	ARG_SIZE            = "-s"
)

// Значения ключей ffmpeg
const (
	VAL_AUDIO_SYNC    = "aresample=async=1:first_pts=0"
	VAL_CODEC_AAC     = "aac"
	VAL_CODEC_H264    = "libx264"
	VAL_CODEC_PCM     = "pcm_s16le"
	VAL_COLOR_UNKNOWN = "unknown"
	VAL_CROP_EVEN     = "crop=trunc(iw/2)*2:trunc(ih/2)*2"
	VAL_FORMAT_PCM    = "s16le"
	VAL_FORMAT_RAW    = "rawvideo"
	VAL_FPS_CONSTANT  = "cfr"
	VAL_FPS_PASS      = "passthrough"
	VAL_LOG_ERROR     = "error"
	VAL_MAP_AUDIO     = "0:a:0"
	VAL_MAP_AUDIO_RAW = "1:a:0"
	VAL_MAP_VIDEO     = "0:v:0"
	VAL_MOV_FAST      = "+faststart"
	VAL_PIPE_AUDIO    = "pipe:3"
	VAL_PIPE_INPUT    = "pipe:0"
	VAL_PIPE_OUTPUT   = "pipe:1"
	VAL_PIX_GRAY      = "gray"
	VAL_PIX_YUV       = "yuv420p"
	VAL_PRESET        = "medium"
	VAL_RANGE_LIMITED = "tv"
)

// Форматы
const FMT_SIZE_FRAME = "%dx%d"

// Размеры
const (
	SIZE_SAMPLE = 2
	SIZE_TAIL   = 2048
)

// Количества
const (
	CNT_BITRATE_CHANNEL = 96000
	CNT_BUFFERS_EXTRA   = 2
	CNT_BUFFERS_WORKER  = 2
)
