package ffpipe

import (
	"context"
	"io"
	"os/exec"

	"stegano2/rescode"
)

// Pass — один проход декодера по видео.
// Args — команда декодера; SizeFrame —
// размер кадра в байтах; IndFirst — номер
// первого кадра прохода; CntFrames — точное
// число кадров прохода (0 — до конца видео).
type Pass struct {
	Args      []string
	SizeFrame int
	IndFirst  int
	CntFrames int
}

// isFull сообщает, набрано ли заданное
// число кадров прохода.
func (pass Pass) isFull(cntFrames int) bool {
	return pass.CntFrames > 0 &&
		cntFrames >= pass.CntFrames
}

// readFrames читает кадры из канала
// процесса по одному и отправляет их
// потребителям.
//
// Параметры:
//   - ctx: контекст отмены.
//   - source: канал вывода декодера.
//   - pass: описание прохода.
//   - lane: каналы конвейера.
//   - last: копия последнего кадра.
//
// Возвращает: число отправленных кадров и
// признак того, что видео закончилось.
func readFrames(
	ctx context.Context,
	source io.Reader,
	pass Pass,
	lane Lane,
	last []byte,
) (int, bool) {
	cntFrames := 0
	isAlive, isEnded := true, false
	for isAlive && !isEnded && !pass.isFull(cntFrames) {
		buffer, isTaken := takeBuffer(ctx, lane)
		if isTaken {
			_, errRead := io.ReadFull(source, buffer)
			isEnded = errRead != nil
		}
		if isTaken && isEnded {
			lane.Pool <- buffer
		}
		isAlive = isTaken && !isEnded
		if isAlive {
			copy(last, buffer)
			isAlive = sendFrame(ctx, lane, Frame{
				Ind: pass.IndFirst + cntFrames,
				Pix: buffer,
			})
			cntFrames++
		}
	}
	return cntFrames, isEnded
}

// padFrames дополняет проход повторами
// последнего кадра до заданного числа.
//
// Параметры:
//   - ctx: контекст отмены.
//   - pass: описание прохода.
//   - lane: каналы конвейера.
//   - last: копия последнего кадра.
//   - cntFrames: число уже отправленных
//     кадров прохода.
//
// Возвращает: число кадров после дополнения.
func padFrames(
	ctx context.Context,
	pass Pass,
	lane Lane,
	last []byte,
	cntFrames int,
) int {
	isAlive := cntFrames > 0
	for isAlive && cntFrames < pass.CntFrames {
		buffer, isTaken := takeBuffer(ctx, lane)
		isAlive = isTaken
		if isAlive {
			copy(buffer, last)
			isAlive = sendFrame(ctx, lane, Frame{
				Ind: pass.IndFirst + cntFrames,
				Pix: buffer,
			})
			cntFrames++
		}
	}
	return cntFrames
}

// ReadPass выполняет один проход декодера:
// запускает ffmpeg и по одному отправляет
// кадры в конвейер. Если задано число
// кадров, лишние отбрасываются, а нехватка
// восполняется повтором последнего кадра.
//
// Параметры:
//   - ctx: контекст отмены.
//   - pass: описание прохода.
//   - lane: каналы конвейера.
//
// Возвращает: число отправленных кадров,
// пояснение ffmpeg и код результата: OK,
// ERR_PIPE_CREATE, ERR_PROC_START либо
// ERR_PROC_DECODE.
func ReadPass(
	ctx context.Context,
	pass Pass,
	lane Lane,
) (int, string, uint8) {
	var tail tailBuffer
	resCode := rescode.OK
	cntFrames := 0
	ctxPass, cancelPass := context.WithCancel(ctx)
	defer cancelPass()
	command := exec.CommandContext(
		ctxPass, NAME_FFMPEG, pass.Args...,
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
		last := make([]byte, pass.SizeFrame)
		isEnded := false
		cntFrames, isEnded = readFrames(
			ctx, source, pass, lane, last,
		)
		if !isEnded {
			cancelPass()
		}
		errWait := command.Wait()
		isFailed := isEnded && ctx.Err() == nil &&
			(errWait != nil || cntFrames == 0)
		if isFailed {
			resCode = rescode.ERR_PROC_DECODE
		}
		if resCode == rescode.OK {
			cntFrames = padFrames(
				ctx, pass, lane, last, cntFrames,
			)
		}
	}
	return cntFrames, tail.text(), resCode
}
