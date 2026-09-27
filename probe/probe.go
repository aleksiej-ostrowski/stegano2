//============================================//
// Пакет probe                                //
//                                            //
// Назначение:                                //
//   анализ видеофайла программой ffprobe:    //
//   размер кадра, частота кадров, число      //
//   кадров, параметры звуковой дорожки.      //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package probe

import (
	"context"
	"encoding/json"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"stegano2/rescode"
)

// Info — сведения о видеофайле.
// Width, Height — размер кодированного кадра;
// IsTurned — кадр повёрнут на четверть
// оборота признаком в контейнере;
// Rate — частота кадров дробью «a/b»;
// CntFrames — число кадров видеодорожки;
// RateAudio, CntChannels — частота выборок и
// число каналов звука (при HasAudio).
type Info struct {
	Width          int
	Height         int
	IsTurned       bool
	Rate           string
	RateNum        int
	RateDen        int
	CntFrames      int
	HasAudio       bool
	RateAudio      int
	CntChannels    int
	ColorSpace     string
	ColorPrimaries string
	ColorTransfer  string
}

// sideData — дополнительные данные дорожки.
type sideData struct {
	Rotation float64 `json:"rotation"`
}

// stream — дорожка в отчёте ffprobe.
type stream struct {
	CodecType      string     `json:"codec_type"`
	Width          int        `json:"width"`
	Height         int        `json:"height"`
	RateAvg        string     `json:"avg_frame_rate"`
	RateBase       string     `json:"r_frame_rate"`
	CntFrames      string     `json:"nb_frames"`
	Duration       string     `json:"duration"`
	RateAudio      string     `json:"sample_rate"`
	CntChannels    int        `json:"channels"`
	ColorSpace     string     `json:"color_space"`
	ColorPrimaries string     `json:"color_primaries"`
	ColorTransfer  string     `json:"color_transfer"`
	SideData       []sideData `json:"side_data_list"`
}

// format — контейнер в отчёте ffprobe.
type format struct {
	Duration string `json:"duration"`
}

// listing — отчёт ffprobe целиком.
type listing struct {
	Streams []stream `json:"streams"`
	Format  format   `json:"format"`
}

// findStream ищет первую дорожку типа.
//
// Параметры:
//   - streams: дорожки файла.
//   - codecType: искомый тип дорожки.
//
// Возвращает: дорожку и признак находки.
func findStream(
	streams []stream,
	codecType string,
) (stream, bool) {
	var found stream
	isFound := false
	indStream := 0
	for !isFound && indStream < len(streams) {
		found = streams[indStream]
		isFound = found.CodecType == codecType
		indStream++
	}
	return found, isFound
}

// parseRate разбирает частоту вида «a/b».
//
// Параметры:
//   - text: запись частоты.
//
// Возвращает: числитель, знаменатель и
// признак того, что оба положительны.
func parseRate(text string) (int, int, bool) {
	parts := strings.Split(text, VAL_RATE_SPLIT)
	rateNum, rateDen := 0, 0
	if len(parts) == CNT_RATE_PARTS {
		rateNum, _ = strconv.Atoi(parts[0])
		rateDen, _ = strconv.Atoi(parts[1])
	}
	return rateNum, rateDen,
		rateNum > 0 && rateDen > 0
}

// fillRate выбирает частоту кадров: среднюю,
// а если её нет — базовую.
//
// Параметры:
//   - info: сведения, дополняются частотой.
//   - video: видеодорожка.
//
// Возвращает: признак того, что частота
// определена.
func fillRate(info *Info, video stream) bool {
	isValid := false
	rates := []string{video.RateAvg, video.RateBase}
	indRate := 0
	for !isValid && indRate < len(rates) {
		info.Rate = rates[indRate]
		info.RateNum, info.RateDen, isValid =
			parseRate(info.Rate)
		indRate++
	}
	return isValid
}

// cntFramesOf определяет число кадров:
// по счётчику контейнера, а если его нет —
// по длительности и частоте кадров.
//
// Параметры:
//   - info: сведения с частотой кадров.
//   - video: видеодорожка.
//   - container: контейнер файла.
//
// Возвращает: число кадров, >= 0.
func cntFramesOf(
	info Info,
	video stream,
	container format,
) int {
	cntFrames, _ := strconv.Atoi(video.CntFrames)
	durations := []string{
		video.Duration,
		container.Duration,
	}
	indTry := 0
	for cntFrames <= 0 && indTry < len(durations) {
		seconds, _ := strconv.ParseFloat(
			durations[indTry], CNT_BITS_FLOAT,
		)
		frames := seconds * float64(info.RateNum) /
			float64(info.RateDen)
		cntFrames = int(frames + LIM_ROUND_HALF)
		indTry++
	}
	return cntFrames
}

// isTurnedOf сообщает, повёрнут ли кадр на
// нечётное число четвертей оборота.
//
// Параметры:
//   - video: видеодорожка.
//
// Возвращает: true для поворота на 90
// или 270 градусов.
func isTurnedOf(video stream) bool {
	isTurned := false
	for _, side := range video.SideData {
		turn := int(math.Abs(side.Rotation))
		isQuart := turn%CNT_TURN_HALF ==
			CNT_TURN_QUART
		isTurned = isTurned != isQuart
	}
	return isTurned
}

// Parse разбирает отчёт ffprobe.
//
// Параметры:
//   - raw: отчёт в формате JSON.
//
// Возвращает: сведения о видео и код
// результата: OK, ERR_PROBE_PARSE,
// ERR_PROBE_VIDEO, ERR_PROBE_SIZE
// либо ERR_PROBE_RATE.
func Parse(raw []byte) (Info, uint8) {
	var info Info
	var parsed listing
	resCode := rescode.OK
	if json.Unmarshal(raw, &parsed) != nil {
		resCode = rescode.ERR_PROBE_PARSE
	}
	video, hasVideo := findStream(
		parsed.Streams, VAL_TYPE_VIDEO,
	)
	if resCode == rescode.OK && !hasVideo {
		resCode = rescode.ERR_PROBE_VIDEO
	}
	if resCode == rescode.OK &&
		(video.Width <= 0 || video.Height <= 0) {
		resCode = rescode.ERR_PROBE_SIZE
	}
	if resCode == rescode.OK &&
		!fillRate(&info, video) {
		resCode = rescode.ERR_PROBE_RATE
	}
	if resCode == rescode.OK {
		audio, hasAudio := findStream(
			parsed.Streams, VAL_TYPE_AUDIO,
		)
		info.Width = video.Width
		info.Height = video.Height
		info.IsTurned = isTurnedOf(video)
		info.CntFrames = cntFramesOf(
			info, video, parsed.Format,
		)
		info.ColorSpace = video.ColorSpace
		info.ColorPrimaries = video.ColorPrimaries
		info.ColorTransfer = video.ColorTransfer
		info.RateAudio, _ = strconv.Atoi(
			audio.RateAudio,
		)
		info.CntChannels = audio.CntChannels
		info.HasAudio = hasAudio &&
			info.RateAudio > 0 &&
			info.CntChannels > 0
	}
	return info, resCode
}

// Probe запускает ffprobe и разбирает отчёт.
//
// Параметры:
//   - ctx: контекст отмены.
//   - path: путь к видеофайлу.
//
// Возвращает: сведения о видео, пояснение
// ffprobe при неудаче и код результата:
// OK, ERR_PROBE_RUN либо код из Parse.
func Probe(
	ctx context.Context,
	path string,
) (Info, string, uint8) {
	var info Info
	var detail strings.Builder
	resCode := rescode.OK
	command := exec.CommandContext(
		ctx,
		NAME_FFPROBE,
		ARG_LOG_LEVEL, VAL_LOG_ERROR,
		ARG_FORMAT, VAL_FORMAT_JSON,
		ARG_SHOW_FORMAT,
		ARG_SHOW_STREAM,
		path,
	)
	command.Stderr = &detail
	raw, errRun := command.Output()
	if errRun != nil {
		resCode = rescode.ERR_PROBE_RUN
	}
	if resCode == rescode.OK {
		info, resCode = Parse(raw)
	}
	text := detail.String()
	if len(text) > SIZE_DETAIL_MAX {
		text = text[:SIZE_DETAIL_MAX]
	}
	return info, strings.TrimSpace(text), resCode
}
