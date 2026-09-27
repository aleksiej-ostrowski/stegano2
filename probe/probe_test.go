package probe

import (
	"testing"

	"stegano2/rescode"
)

const (
	JSON_FULL = `{"streams":[
		{"codec_type":"audio","sample_rate":"48000",
		 "channels":2},
		{"codec_type":"video","width":1920,
		 "height":1080,"avg_frame_rate":"30000/1001",
		 "r_frame_rate":"30000/1001","nb_frames":"301",
		 "color_space":"bt709",
		 "side_data_list":[{"rotation":-90}]}],
		"format":{"duration":"10.043367"}}`
	JSON_BARE = `{"streams":[
		{"codec_type":"video","width":640,
		 "height":360,"avg_frame_rate":"0/0",
		 "r_frame_rate":"25/1"}],
		"format":{"duration":"4.04"}}`
	JSON_SILENT = `{"streams":[
		{"codec_type":"audio","sample_rate":"44100",
		 "channels":1}],"format":{}}`
	JSON_BROKEN   = `{"streams":[`
	JSON_RATELESS = `{"streams":[
		{"codec_type":"video","width":8,"height":8}]}`
	JSON_SIZELESS = `{"streams":[
		{"codec_type":"video"}]}`
	RATE_FULL       = "30000/1001"
	RATE_BARE       = "25/1"
	SPACE_FULL      = "bt709"
	SIZE_WIDTH_FULL = 1920
	CNT_FRAMES_FULL = 301
	CNT_FRAMES_BARE = 101
	RATE_AUDIO_FULL = 48000
	CNT_CHAN_FULL   = 2
)

func TestParseFull(test *testing.T) {
	info, resCode := Parse([]byte(JSON_FULL))
	isRight := resCode == rescode.OK &&
		info.Width == SIZE_WIDTH_FULL &&
		info.IsTurned &&
		info.Rate == RATE_FULL &&
		info.CntFrames == CNT_FRAMES_FULL &&
		info.HasAudio &&
		info.RateAudio == RATE_AUDIO_FULL &&
		info.CntChannels == CNT_CHAN_FULL &&
		info.ColorSpace == SPACE_FULL
	if !isRight {
		test.Fatalf("код %d: %+v", resCode, info)
	}
}

func TestParseBare(test *testing.T) {
	info, resCode := Parse([]byte(JSON_BARE))
	isRight := resCode == rescode.OK &&
		info.Rate == RATE_BARE &&
		info.CntFrames == CNT_FRAMES_BARE &&
		!info.HasAudio && !info.IsTurned
	if !isRight {
		test.Fatalf("код %d: %+v", resCode, info)
	}
}

func TestParseErrors(test *testing.T) {
	_, codeBroken := Parse([]byte(JSON_BROKEN))
	_, codeSilent := Parse([]byte(JSON_SILENT))
	_, codeRate := Parse([]byte(JSON_RATELESS))
	_, codeSize := Parse([]byte(JSON_SIZELESS))
	isRight :=
		codeBroken == rescode.ERR_PROBE_PARSE &&
			codeSilent == rescode.ERR_PROBE_VIDEO &&
			codeRate == rescode.ERR_PROBE_RATE &&
			codeSize == rescode.ERR_PROBE_SIZE
	if !isRight {
		test.Fatalf("коды %d %d %d %d", codeBroken,
			codeSilent, codeRate, codeSize)
	}
}
