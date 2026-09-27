package main

import (
	"bytes"
	"context"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"stegano2/ffpipe"
	"stegano2/merge"
	"stegano2/mode"
	"stegano2/probe"
	"stegano2/rescode"
	"stegano2/split"
)

const (
	SEED_TEST        = 20260927
	KEY_TEST         = "ключ"
	KEY_WRONG        = "замок"
	SIZE_DATA_TEST   = 200
	MODE_FILE_TEST   = 0o644
	NAME_DATA        = "data.bin"
	NAME_ORIGINAL    = "original.mp4"
	NAME_SILENT      = "silent.mp4"
	NAME_RESULT      = "result.mp4"
	NAME_OUTPUT      = "output.bin"
	CNT_FRAMES_PASS  = 25
	CNT_PASSES_REF   = 4
	CNT_FILES_REF    = 4
	ARG_LAVFI        = "lavfi"
	ARG_SHORTEST     = "-shortest"
	VAL_SOURCE_VIDEO = "testsrc2=size=320x240" +
		":rate=25:duration=1"
	VAL_SOURCE_AUDIO = "sine=frequency=440" +
		":sample_rate=44100:duration=1"
)

// requireTools пропускает проверку, если
// нет ffmpeg и ffprobe.
func requireTools(test *testing.T) {
	if ffpipe.CheckTools() != rescode.OK {
		test.Skip("нет ffmpeg и ffprobe")
	}
}

// makeData создаёт файл случайных данных.
func makeData(test *testing.T, dir string) []byte {
	generator := rand.New(rand.NewSource(SEED_TEST))
	data := make([]byte, SIZE_DATA_TEST)
	generator.Read(data)
	errWrite := os.WriteFile(
		filepath.Join(dir, NAME_DATA),
		data,
		MODE_FILE_TEST,
	)
	if errWrite != nil {
		test.Fatalf("данные: %v", errWrite)
	}
	return data
}

// makeVideo создаёт короткое видео: со
// звуковой дорожкой или без неё.
func makeVideo(
	test *testing.T,
	path string,
	hasAudio bool,
) {
	args := []string{
		ffpipe.ARG_LOG_LEVEL, ffpipe.VAL_LOG_ERROR,
		ffpipe.ARG_OVERWRITE,
		ffpipe.ARG_FORMAT, ARG_LAVFI,
		ffpipe.ARG_INPUT, VAL_SOURCE_VIDEO,
	}
	if hasAudio {
		args = append(
			args,
			ffpipe.ARG_FORMAT, ARG_LAVFI,
			ffpipe.ARG_INPUT, VAL_SOURCE_AUDIO,
			ARG_SHORTEST,
		)
	}
	args = append(
		args,
		ffpipe.ARG_PIX_FMT, ffpipe.VAL_PIX_YUV,
		path,
	)
	output, errRun := exec.Command(
		ffpipe.NAME_FFMPEG, args...,
	).CombinedOutput()
	if errRun != nil {
		test.Fatalf("видео: %v\n%s", errRun, output)
	}
}

// keepQuiet — извещение о ходе работы,
// которое ничего не делает.
func keepQuiet(cntDone, cntTotal int) {}

// roundTrip скрывает данные в видео и
// извлекает их обратно.
func roundTrip(
	test *testing.T,
	dir string,
	nameOriginal string,
	params mode.Params,
) (probe.Info, []byte) {
	ctx := context.Background()
	pathResult := filepath.Join(dir, NAME_RESULT)
	pathOutput := filepath.Join(dir, NAME_OUTPUT)
	merged := merge.Run(
		ctx,
		merge.Args{
			Key:          KEY_TEST,
			Params:       params,
			PathData:     filepath.Join(dir, NAME_DATA),
			PathOriginal: filepath.Join(dir, nameOriginal),
			PathResult:   pathResult,
		},
		keepQuiet,
	)
	if merged.Code != rescode.OK {
		test.Fatalf("merge: код %d\n%s",
			merged.Code, merged.Detail)
	}
	parted := split.Run(
		ctx,
		split.Args{
			Key:        KEY_TEST,
			PathInput:  pathResult,
			PathOutput: pathOutput,
		},
		keepQuiet,
	)
	if parted.Code != rescode.OK {
		test.Fatalf("split: код %d\n%s",
			parted.Code, parted.Detail)
	}
	isNamed := parted.NameData == NAME_DATA &&
		parted.NameMode == params.Name
	if !isNamed {
		test.Fatalf("отчёт %+v", parted)
	}
	info, _, _ := probe.Probe(ctx, pathResult)
	output, _ := os.ReadFile(pathOutput)
	return info, output
}

func TestRoundTripWithAudio(test *testing.T) {
	requireTools(test)
	dir := test.TempDir()
	data := makeData(test, dir)
	makeVideo(
		test, filepath.Join(dir, NAME_ORIGINAL), true,
	)
	info, output := roundTrip(
		test, dir, NAME_ORIGINAL, mode.MODES[0],
	)
	isRight := bytes.Equal(output, data) &&
		info.HasAudio &&
		info.CntFrames ==
			CNT_PASSES_REF*CNT_FRAMES_PASS
	if !isRight {
		test.Fatalf("видео %+v", info)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != CNT_FILES_REF {
		test.Fatalf("лишние файлы: %v", entries)
	}
}

func TestRoundTripModesSilent(test *testing.T) {
	requireTools(test)
	dir := test.TempDir()
	data := makeData(test, dir)
	makeVideo(
		test, filepath.Join(dir, NAME_SILENT), false,
	)
	for _, params := range mode.MODES {
		info, output := roundTrip(
			test, dir, NAME_SILENT, params,
		)
		isRight := bytes.Equal(output, data) &&
			!info.HasAudio
		if !isRight {
			test.Fatalf("режим %s: видео %+v",
				params.Name, info)
		}
	}
}

func TestSplitWrongKey(test *testing.T) {
	requireTools(test)
	dir := test.TempDir()
	makeData(test, dir)
	makeVideo(
		test, filepath.Join(dir, NAME_SILENT), false,
	)
	roundTrip(test, dir, NAME_SILENT, mode.MODES[0])
	pathOutput := filepath.Join(dir, NAME_OUTPUT)
	os.Remove(pathOutput)
	parted := split.Run(
		context.Background(),
		split.Args{
			Key:        KEY_WRONG,
			PathInput:  filepath.Join(dir, NAME_RESULT),
			PathOutput: pathOutput,
		},
		keepQuiet,
	)
	_, errStat := os.Stat(pathOutput)
	isRight := errStat != nil &&
		parted.Code == rescode.ERR_STREAM_HEADER
	if !isRight {
		test.Fatalf("код %d", parted.Code)
	}
}
