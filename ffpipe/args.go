package ffpipe

import (
	"fmt"
	"strconv"

	"stegano2/probe"
)

// argsCommon возвращает общее начало команды.
func argsCommon() []string {
	return []string{
		ARG_LOG_LEVEL, VAL_LOG_ERROR,
		ARG_NO_STDIN,
	}
}

// ArgsVideoSource строит команду декодера,
// выдающего кадры yuv420p с постоянной
// частотой и чётными размерами.
//
// Параметры:
//   - path: путь к исходному видео.
//   - info: сведения об исходном видео.
//
// Возвращает: аргументы ffmpeg.
func ArgsVideoSource(
	path string,
	info probe.Info,
) []string {
	return append(
		argsCommon(),
		ARG_INPUT, path,
		ARG_MAP, VAL_MAP_VIDEO,
		ARG_NO_AUDIO, ARG_NO_SUBS, ARG_NO_DATA,
		ARG_FILTER_VIDEO, VAL_CROP_EVEN,
		ARG_FPS_MODE, VAL_FPS_CONSTANT,
		ARG_RATE, info.Rate,
		ARG_PIX_FMT, VAL_PIX_YUV,
		ARG_FORMAT, VAL_FORMAT_RAW,
		VAL_PIPE_OUTPUT,
	)
}

// ArgsGraySource строит команду декодера,
// выдающего яркость каждого кадра ровно
// один раз, без поворота.
//
// Параметры:
//   - path: путь к видео со скрытыми данными.
//
// Возвращает: аргументы ffmpeg.
func ArgsGraySource(path string) []string {
	return append(
		argsCommon(),
		ARG_NO_ROTATE,
		ARG_INPUT, path,
		ARG_MAP, VAL_MAP_VIDEO,
		ARG_NO_AUDIO, ARG_NO_SUBS, ARG_NO_DATA,
		ARG_FPS_MODE, VAL_FPS_PASS,
		ARG_PIX_FMT, VAL_PIX_GRAY,
		ARG_FORMAT, VAL_FORMAT_RAW,
		VAL_PIPE_OUTPUT,
	)
}

// ArgsAudioSource строит команду декодера
// звука: выборки s16le, разрывы меток
// времени заполняются тишиной.
//
// Параметры:
//   - path: путь к исходному видео.
//   - info: сведения об исходном видео.
//
// Возвращает: аргументы ffmpeg.
func ArgsAudioSource(
	path string,
	info probe.Info,
) []string {
	return append(
		argsCommon(),
		ARG_INPUT, path,
		ARG_MAP, VAL_MAP_AUDIO,
		ARG_NO_VIDEO, ARG_NO_SUBS, ARG_NO_DATA,
		ARG_FILTER_AUDIO, VAL_AUDIO_SYNC,
		ARG_RATE_AUDIO, strconv.Itoa(info.RateAudio),
		ARG_CHANNELS, strconv.Itoa(info.CntChannels),
		ARG_CODEC_AUDIO, VAL_CODEC_PCM,
		ARG_FORMAT, VAL_FORMAT_PCM,
		VAL_PIPE_OUTPUT,
	)
}

// argsColor добавляет известные признаки
// цветового пространства.
//
// Параметры:
//   - args: накапливаемые аргументы.
//   - info: сведения об исходном видео.
//
// Возвращает: дополненные аргументы.
func argsColor(
	args []string,
	info probe.Info,
) []string {
	pairs := [][]string{
		{ARG_COLOR_SPACE, info.ColorSpace},
		{ARG_COLOR_PRIMARIES, info.ColorPrimaries},
		{ARG_COLOR_TRANSFER, info.ColorTransfer},
	}
	for _, pair := range pairs {
		isKnown := len(pair[1]) > 0 &&
			pair[1] != VAL_COLOR_UNKNOWN
		if isKnown {
			args = append(args, pair...)
		}
	}
	return append(
		args,
		ARG_COLOR_RANGE, VAL_RANGE_LIMITED,
	)
}

// argsAudioInput добавляет вход звука из
// дополнительного канала процесса.
func argsAudioInput(
	args []string,
	info probe.Info,
) []string {
	return append(
		args,
		ARG_FORMAT, VAL_FORMAT_PCM,
		ARG_RATE_AUDIO, strconv.Itoa(info.RateAudio),
		ARG_CHANNELS, strconv.Itoa(info.CntChannels),
		ARG_INPUT, VAL_PIPE_AUDIO,
	)
}

// argsAudioOutput добавляет дорожку звука
// и её кодирование в AAC.
func argsAudioOutput(
	args []string,
	info probe.Info,
) []string {
	bitrate := info.CntChannels * CNT_BITRATE_CHANNEL
	return append(
		args,
		ARG_MAP, VAL_MAP_AUDIO_RAW,
		ARG_CODEC_AUDIO, VAL_CODEC_AAC,
		ARG_BITRATE_AUDIO, strconv.Itoa(bitrate),
	)
}

// ArgsEncoder строит команду кодировщика:
// кадры yuv420p приходят на стандартный
// вход, звук — по дополнительному каналу.
//
// Параметры:
//   - path: путь к создаваемому видео.
//   - width, height: размер кадра, чётные.
//   - info: сведения об исходном видео.
//   - crf: качество сжатия H.264.
//
// Возвращает: аргументы ffmpeg.
func ArgsEncoder(
	path string,
	width, height int,
	info probe.Info,
	crf string,
) []string {
	size := fmt.Sprintf(FMT_SIZE_FRAME, width, height)
	args := append(
		argsCommon(),
		ARG_OVERWRITE,
		ARG_FORMAT, VAL_FORMAT_RAW,
		ARG_PIX_FMT, VAL_PIX_YUV,
		ARG_SIZE, size,
		ARG_RATE, info.Rate,
		ARG_INPUT, VAL_PIPE_INPUT,
	)
	if info.HasAudio {
		args = argsAudioInput(args, info)
	}
	args = append(
		args,
		ARG_MAP, VAL_MAP_VIDEO,
		ARG_CODEC_VIDEO, VAL_CODEC_H264,
		ARG_PRESET, VAL_PRESET,
		ARG_CRF, crf,
		ARG_PIX_FMT, VAL_PIX_YUV,
	)
	args = argsColor(args, info)
	if info.HasAudio {
		args = argsAudioOutput(args, info)
	}
	return append(
		args,
		ARG_MOV_FLAGS, VAL_MOV_FAST,
		path,
	)
}
