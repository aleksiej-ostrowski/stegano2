package merge

import (
	"testing"

	"stegano2/probe"
	"stegano2/rescode"
)

const (
	SIZE_WIDTH_TEST   = 1920
	SIZE_HEIGHT_TEST  = 1080
	SIZE_WIDTH_ODD    = 853
	SIZE_HEIGHT_ODD   = 479
	SIZE_WIDTH_EVEN   = 852
	SIZE_HEIGHT_EVEN  = 478
	SIZE_TINY         = 6
	RATE_NUM_TEST     = 30000
	RATE_DEN_TEST     = 1001
	RATE_AUDIO_TEST   = 48000
	CNT_CHANNELS_TEST = 2
	CNT_FRAMES_TEST   = 301
	SIZE_STREAM_TEST  = 7305888
	CNT_PASSES_REF    = 6
	CNT_FRAMES_REF    = 1806
	CNT_SAMPLES_REF   = 2892489
	CNT_PASSES_LONG   = 1000
	SIZE_FRAME_AUDIO  = 4
	SEED_TEST         = 12345
	CNT_BITS_TEST     = 100000
	CNT_ONES_LOW      = 49000
	CNT_ONES_HIGH     = 51000
)

// sampleInfo возвращает сведения об образце.
func sampleInfo() probe.Info {
	return probe.Info{
		Width:       SIZE_WIDTH_TEST,
		Height:      SIZE_HEIGHT_TEST,
		RateNum:     RATE_NUM_TEST,
		RateDen:     RATE_DEN_TEST,
		CntFrames:   CNT_FRAMES_TEST,
		HasAudio:    true,
		RateAudio:   RATE_AUDIO_TEST,
		CntChannels: CNT_CHANNELS_TEST,
	}
}

func TestMakePlan(test *testing.T) {
	plan, resCode := MakePlan(
		sampleInfo(), SIZE_STREAM_TEST,
	)
	isRight := resCode == rescode.OK &&
		plan.CntPasses == CNT_PASSES_REF &&
		plan.CntFrames == CNT_FRAMES_REF &&
		plan.CntFramesPass == CNT_FRAMES_TEST
	if !isRight {
		test.Fatalf("код %d: %+v", resCode, plan)
	}
}

func TestMakePlanShapes(test *testing.T) {
	info := sampleInfo()
	info.Width = SIZE_HEIGHT_ODD
	info.Height = SIZE_WIDTH_ODD
	info.IsTurned = true
	plan, resCode := MakePlan(info, 1)
	isRight := resCode == rescode.OK &&
		plan.Grid.Width == SIZE_WIDTH_EVEN &&
		plan.Grid.Height == SIZE_HEIGHT_EVEN &&
		plan.CntPasses == 1
	if !isRight {
		test.Fatalf("код %d: %+v", resCode, plan)
	}
	info.Width = SIZE_TINY
	_, codeTiny := MakePlan(info, 1)
	info = sampleInfo()
	info.CntFrames = 0
	_, codeEmpty := MakePlan(info, 1)
	if codeTiny != rescode.ERR_PROBE_SIZE ||
		codeEmpty != rescode.ERR_PROBE_SIZE {
		test.Fatalf("коды %d %d", codeTiny, codeEmpty)
	}
}

func TestAudioMatchesVideo(test *testing.T) {
	info := sampleInfo()
	plan, _ := MakePlan(info, SIZE_STREAM_TEST)
	sizeTotal := int64(0)
	for indPass := 0; indPass < plan.CntPasses; indPass++ {
		sizePass := SizeAudioPass(info, plan, indPass)
		if sizePass%SIZE_FRAME_AUDIO != 0 {
			test.Fatalf("проход %d: дробная выборка",
				indPass)
		}
		sizeTotal += sizePass
	}
	if sizeTotal != CNT_SAMPLES_REF*SIZE_FRAME_AUDIO {
		test.Fatalf("байтов звука %d", sizeTotal)
	}
}

func TestAudioHasNoDrift(test *testing.T) {
	info := sampleInfo()
	plan := Plan{CntFramesPass: CNT_FRAMES_TEST}
	sizeTotal := int64(0)
	for indPass := 0; indPass < CNT_PASSES_LONG; indPass++ {
		sizeTotal += SizeAudioPass(info, plan, indPass)
	}
	cntSamples := cntSamplesAt(
		info, CNT_PASSES_LONG*CNT_FRAMES_TEST,
	)
	if sizeTotal != cntSamples*SIZE_FRAME_AUDIO {
		test.Fatalf("звук разошёлся с видео")
	}
}

func TestDustBit(test *testing.T) {
	cntOnes := 0
	for indBit := 0; indBit < CNT_BITS_TEST; indBit++ {
		bit := DustBit(SEED_TEST, indBit)
		if bit != DustBit(SEED_TEST, indBit) {
			test.Fatalf("бит %d непостоянен", indBit)
		}
		cntOnes += int(bit)
	}
	isFair := cntOnes > CNT_ONES_LOW &&
		cntOnes < CNT_ONES_HIGH
	if !isFair {
		test.Fatalf("единиц %d", cntOnes)
	}
}
