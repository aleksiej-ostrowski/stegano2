package ffpipe

import (
	"context"
	"io"
	"os/exec"

	"stegano2/rescode"
)

// silence — бесконечный источник тишины.
type silence struct{}

// Read заполняет буфер нулевыми выборками.
func (silence) Read(chunk []byte) (int, error) {
	clear(chunk)
	return len(chunk), nil
}

// PumpAudio выполняет один проход декодера
// звука и передаёт кодировщику ровно
// заданное число байтов выборок: лишний
// звук отбрасывается, нехватка восполняется
// тишиной. Так длительность звука каждого
// прохода совпадает с длительностью видео.
//
// Параметры:
//   - ctx: контекст отмены.
//   - args: команда декодера звука.
//   - cntBytes: число байтов прохода, >= 0.
//   - target: канал звука кодировщика.
//
// Возвращает: пояснение ffmpeg и код
// результата: OK, ERR_PIPE_CREATE,
// ERR_PROC_START, ERR_PROC_AUDIO либо
// ERR_PIPE_WRITE.
func PumpAudio(
	ctx context.Context,
	args []string,
	cntBytes int64,
	target io.Writer,
) (string, uint8) {
	var tail tailBuffer
	resCode := rescode.OK
	ctxPass, cancelPass := context.WithCancel(ctx)
	defer cancelPass()
	command := exec.CommandContext(
		ctxPass, NAME_FFMPEG, args...,
	)
	command.Stderr = &tail
	source, errPipe := command.StdoutPipe()
	if errPipe != nil {
		resCode = rescode.ERR_PIPE_CREATE
	}
	if resCode == rescode.OK &&
		command.Start() != nil {
		resCode = rescode.ERR_PROC_START
	}
	if resCode == rescode.OK {
		cntCopied, errCopy := io.CopyN(
			target, source, cntBytes,
		)
		isEnded := errCopy == io.EOF
		if !isEnded {
			cancelPass()
		}
		errWait := command.Wait()
		isAlive := ctx.Err() == nil
		if isAlive && isEnded && errWait != nil {
			resCode = rescode.ERR_PROC_AUDIO
		}
		if isAlive && !isEnded && errCopy != nil {
			resCode = rescode.ERR_PIPE_WRITE
		}
		if isAlive && resCode == rescode.OK {
			_, errPad := io.CopyN(
				target,
				silence{},
				cntBytes-cntCopied,
			)
			if errPad != nil {
				resCode = rescode.ERR_PIPE_WRITE
			}
		}
	}
	return tail.text(), resCode
}
