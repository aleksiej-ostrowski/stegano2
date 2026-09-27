//============================================//
// Пакет ffpipe                               //
//                                            //
// Назначение:                                //
//   конвейер кадров между процессами ffmpeg  //
//   и потребителями. Видео не распаковыва-   //
//   ется на диск: декодер пишет кадры в      //
//   канал процесса, читатель отправляет их   //
//   по одному в канал Go, потребители раз-   //
//   бирают кадры параллельно, упорядочива-   //
//   тель возвращает исходный порядок.        //
//                                            //
//   декодер -> Frames -> потребители ->      //
//           -> Done -> упорядочиватель       //
//                                            //
//   Буферы кадров ходят по кругу через       //
//   канал Pool, поэтому память ограничена    //
//   числом буферов.                          //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package ffpipe

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"stegano2/rescode"
)

// Frame — кадр, идущий по конвейеру.
// Ind — порядковый номер кадра; Pix —
// пиксели; Bits — биты, распознанные
// потребителем (при извлечении данных).
type Frame struct {
	Ind  int
	Pix  []byte
	Bits []uint8
}

// Lane — каналы конвейера кадров.
// Pool — свободные буферы; Frames — кадры
// для потребителей; Done — обработанные.
type Lane struct {
	Pool   chan []byte
	Frames chan Frame
	Done   chan Frame
}

// tailBuffer хранит конец потока ошибок
// внешней программы.
type tailBuffer struct {
	data []byte
}

// Write дописывает данные, сохраняя только
// последние SIZE_TAIL байтов.
func (tail *tailBuffer) Write(
	chunk []byte,
) (int, error) {
	tail.data = append(tail.data, chunk...)
	cntExtra := len(tail.data) - SIZE_TAIL
	if cntExtra > 0 {
		tail.data = tail.data[cntExtra:]
	}
	return len(chunk), nil
}

// text возвращает накопленный текст.
func (tail *tailBuffer) text() string {
	return strings.TrimSpace(string(tail.data))
}

// CntWorkers определяет число потребителей.
//
// Возвращает: число потребителей, >= 1.
func CntWorkers() int {
	return max(1, runtime.NumCPU())
}

// NewLane создаёт каналы конвейера и буферы.
//
// Параметры:
//   - cntWorkers: число потребителей, >= 1.
//   - sizeFrame: размер кадра в байтах.
//
// Возвращает: конвейер с полным запасом
// свободных буферов.
func NewLane(cntWorkers, sizeFrame int) Lane {
	cntBuffers := cntWorkers*CNT_BUFFERS_WORKER +
		CNT_BUFFERS_EXTRA
	lane := Lane{
		Pool:   make(chan []byte, cntBuffers),
		Frames: make(chan Frame, cntWorkers),
		Done:   make(chan Frame, cntBuffers),
	}
	for indBuffer := 0; indBuffer < cntBuffers; indBuffer++ {
		lane.Pool <- make([]byte, sizeFrame)
	}
	return lane
}

// CheckTools проверяет наличие ffmpeg
// и ffprobe.
//
// Возвращает: OK либо ERR_PROC_LOOKUP.
func CheckTools() uint8 {
	resCode := rescode.OK
	names := []string{NAME_FFMPEG, NAME_FFPROBE}
	for _, name := range names {
		_, errLook := exec.LookPath(name)
		if errLook != nil {
			resCode = rescode.ERR_PROC_LOOKUP
		}
	}
	return resCode
}

// takeBuffer берёт свободный буфер, ожидая
// его возврата из конвейера.
//
// Возвращает: буфер и признак успеха; при
// отмене контекста — неуспех.
func takeBuffer(
	ctx context.Context,
	lane Lane,
) ([]byte, bool) {
	var buffer []byte
	isTaken := false
	select {
	case <-ctx.Done():
	case buffer = <-lane.Pool:
		isTaken = true
	}
	return buffer, isTaken
}

// sendFrame отправляет кадр потребителям.
//
// Возвращает: признак успеха; при отмене
// контекста — неуспех.
func sendFrame(
	ctx context.Context,
	lane Lane,
	frame Frame,
) bool {
	isSent := false
	select {
	case <-ctx.Done():
	case lane.Frames <- frame:
		isSent = true
	}
	return isSent
}

// RunWorkers запускает потребителей кадров
// и ждёт, пока канал Frames не опустеет и
// не закроется; затем закрывает канал Done.
//
// Параметры:
//   - ctx: контекст отмены; после отмены
//     кадры пропускаются без обработки.
//   - cntWorkers: число потребителей, >= 1.
//   - lane: каналы конвейера.
//   - transform: обработка одного кадра.
func RunWorkers(
	ctx context.Context,
	cntWorkers int,
	lane Lane,
	transform func(frame *Frame),
) {
	var group sync.WaitGroup
	for indWorker := 0; indWorker < cntWorkers; indWorker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for frame := range lane.Frames {
				if ctx.Err() == nil {
					transform(&frame)
				}
				lane.Done <- frame
			}
		}()
	}
	group.Wait()
	close(lane.Done)
}

// OrderFrames принимает обработанные кадры,
// восстанавливает их порядок и передаёт
// обработчику; буфер кадра возвращается
// в запас. Работает до закрытия Done.
//
// Параметры:
//   - lane: каналы конвейера.
//   - handle: обработчик очередного кадра.
func OrderFrames(
	lane Lane,
	handle func(frame Frame),
) {
	pending := make(map[int]Frame)
	indNext := 0
	for frame := range lane.Done {
		pending[frame.Ind] = frame
		next, isReady := pending[indNext]
		for isReady {
			delete(pending, indNext)
			handle(next)
			lane.Pool <- next.Pix
			indNext++
			next, isReady = pending[indNext]
		}
	}
}
