package ffpipe

import (
	"context"
	"io"
	"os"
	"os/exec"

	"stegano2/rescode"
)

// Encoder — работающий кодировщик ffmpeg.
// Video — канал кадров (стандартный вход);
// Audio — канал звука (nil, если звука нет).
type Encoder struct {
	Video   io.WriteCloser
	Audio   io.WriteCloser
	command *exec.Cmd
	tail    tailBuffer
}

// StartEncoder запускает кодировщик.
//
// Параметры:
//   - ctx: контекст отмены.
//   - args: команда кодировщика.
//   - hasAudio: нужен ли канал звука.
//
// Возвращает: кодировщик и код результата:
// OK, ERR_PIPE_CREATE либо ERR_PROC_START.
func StartEncoder(
	ctx context.Context,
	args []string,
	hasAudio bool,
) (*Encoder, uint8) {
	var audioInput, audioOutput *os.File
	var errAudio error
	encoder := &Encoder{}
	resCode := rescode.OK
	encoder.command = exec.CommandContext(
		ctx, NAME_FFMPEG, args...,
	)
	encoder.command.Stderr = &encoder.tail
	video, errVideo := encoder.command.StdinPipe()
	if hasAudio {
		audioInput, audioOutput, errAudio = os.Pipe()
	}
	if errVideo != nil || errAudio != nil {
		resCode = rescode.ERR_PIPE_CREATE
	}
	if resCode == rescode.OK && hasAudio {
		encoder.command.ExtraFiles =
			[]*os.File{audioInput}
		encoder.Audio = audioOutput
	}
	if resCode == rescode.OK {
		encoder.Video = video
		if encoder.command.Start() != nil {
			resCode = rescode.ERR_PROC_START
		}
	}
	if audioInput != nil {
		audioInput.Close()
	}
	if resCode != rescode.OK && audioOutput != nil {
		audioOutput.Close()
	}
	return encoder, resCode
}

// Finish ждёт завершения кодировщика.
// Каналы Video и Audio к этому моменту
// должны быть закрыты.
//
// Возвращает: пояснение ffmpeg и код
// результата: OK либо ERR_PROC_ENCODE.
func (encoder *Encoder) Finish() (string, uint8) {
	resCode := rescode.OK
	if encoder.command.Wait() != nil {
		resCode = rescode.ERR_PROC_ENCODE
	}
	return encoder.tail.text(), resCode
}
