package ffpipe

import (
	"context"
	"slices"
	"strings"
	"testing"

	"stegano2/probe"
)

const (
	CNT_WORKERS_TEST = 4
	CNT_FRAMES_TEST  = 200
	SIZE_FRAME_TEST  = 16
	CNT_BUFFERS_REF  = 10
	SIZE_WRITE_TEST  = 1500
	TEXT_TAIL_MARK   = "конец"
	PATH_TEST        = "video.mp4"
	RATE_TEST        = "25/1"
	SPACE_TEST       = "bt709"
	CRF_TEST         = "18"
	SIZE_WIDTH_TEST  = 320
	SIZE_HEIGHT_TEST = 240
	VAL_SIZE_REF     = "320x240"
)

// feedFrames отправляет в конвейер кадры,
// помеченные своим номером.
func feedFrames(ctx context.Context, lane Lane) {
	for indFrame := 0; indFrame < CNT_FRAMES_TEST; indFrame++ {
		buffer, _ := takeBuffer(ctx, lane)
		buffer[0] = uint8(indFrame)
		sendFrame(ctx, lane, Frame{
			Ind: indFrame,
			Pix: buffer,
		})
	}
	close(lane.Frames)
}

func TestLaneKeepsOrder(test *testing.T) {
	ctx := context.Background()
	lane := NewLane(CNT_WORKERS_TEST, SIZE_FRAME_TEST)
	if cap(lane.Pool) != CNT_BUFFERS_REF {
		test.Fatalf("буферов %d", cap(lane.Pool))
	}
	go feedFrames(ctx, lane)
	go RunWorkers(
		ctx, CNT_WORKERS_TEST, lane,
		func(frame *Frame) {
			frame.Pix[1] = ^frame.Pix[0]
		},
	)
	cntSeen := 0
	OrderFrames(lane, func(frame Frame) {
		isRight := frame.Ind == cntSeen &&
			frame.Pix[0] == uint8(cntSeen) &&
			frame.Pix[1] == ^uint8(cntSeen)
		if !isRight {
			test.Errorf("кадр %d на месте %d",
				frame.Ind, cntSeen)
		}
		cntSeen++
	})
	if cntSeen != CNT_FRAMES_TEST {
		test.Fatalf("кадров %d", cntSeen)
	}
}

func TestLaneStopsOnCancel(test *testing.T) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	lane := NewLane(CNT_WORKERS_TEST, SIZE_FRAME_TEST)
	for range cap(lane.Pool) {
		takeBuffer(ctx, lane)
	}
	for indFrame := 0; indFrame < CNT_WORKERS_TEST; indFrame++ {
		sendFrame(ctx, lane, Frame{Ind: indFrame})
	}
	cancel()
	_, isTaken := takeBuffer(ctx, lane)
	isSent := sendFrame(ctx, lane, Frame{})
	if isTaken || isSent {
		test.Fatalf("работа после отмены: %v %v",
			isTaken, isSent)
	}
}

func TestTailBuffer(test *testing.T) {
	var tail tailBuffer
	filler := strings.Repeat(RATE_TEST, SIZE_WRITE_TEST)
	tail.Write([]byte(filler))
	tail.Write([]byte(TEXT_TAIL_MARK + "\n"))
	text := tail.text()
	isRight := len(tail.data) == SIZE_TAIL &&
		strings.HasSuffix(text, TEXT_TAIL_MARK)
	if !isRight {
		test.Fatalf("хвост длиной %d", len(tail.data))
	}
}

func TestArgsEncoder(test *testing.T) {
	info := probe.Info{
		Rate:        RATE_TEST,
		HasAudio:    true,
		RateAudio:   CNT_BITRATE_CHANNEL,
		CntChannels: SIZE_SAMPLE,
		ColorSpace:  SPACE_TEST,
	}
	args := ArgsEncoder(
		PATH_TEST,
		SIZE_WIDTH_TEST,
		SIZE_HEIGHT_TEST,
		info,
		CRF_TEST,
	)
	wanted := []string{
		VAL_SIZE_REF, VAL_PIPE_INPUT,
		VAL_PIPE_AUDIO, VAL_MAP_AUDIO_RAW,
		ARG_COLOR_SPACE, VAL_CODEC_AAC,
	}
	for _, value := range wanted {
		if !slices.Contains(args, value) {
			test.Fatalf("нет аргумента %s", value)
		}
	}
	isRight := args[len(args)-1] == PATH_TEST &&
		!slices.Contains(args, ARG_COLOR_TRANSFER)
	if !isRight {
		test.Fatalf("аргументы %v", args)
	}
	info.HasAudio = false
	silent := ArgsEncoder(
		PATH_TEST,
		SIZE_WIDTH_TEST,
		SIZE_HEIGHT_TEST,
		info,
		CRF_TEST,
	)
	if slices.Contains(silent, VAL_PIPE_AUDIO) {
		test.Fatalf("звук без дорожки: %v", silent)
	}
}
