//============================================//
// Пакет split                                //
//                                            //
// Назначение:                                //
//   команда split — извлечение скрытого      //
//   файла данных из видео.                   //
//                                            //
// Входные данные:                            //
//   ключ, видео со скрытыми данными          //
// Выходные данные:                           //
//   файл данных                              //
//                                            //
// Порядок работы:                            //
//   1. декодер отдаёт яркость кадров по      //
//      одному; потребители распознают биты   //
//      ячеек; на диск кадры не распаковы-    //
//      ваются, звук пропускается;            //
//   2. по заголовку определяются режим и     //
//      размер потока; чтение видео           //
//      прекращается, как только набран       //
//      весь поток;                           //
//   3. поток -> обратное перемешивание ->    //
//      голосование копий -> БЧХ -> данные,   //
//      сверка контрольной суммы.             //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package split

import (
	"context"
	"math"
	"os"
	"sync"

	"stegano2/cell"
	"stegano2/ffpipe"
	"stegano2/payload"
	"stegano2/probe"
	"stegano2/report"
	"stegano2/rescode"
)

// Args — аргументы команды split.
type Args struct {
	Key        string
	PathInput  string
	PathOutput string
}

// job — состояние выполняемой команды.
type job struct {
	ctx        context.Context
	args       Args
	onProgress func(cntDone, cntTotal int)
	info       probe.Info
	grid       cell.Grid
	stream     []byte
	data       []byte
	result     report.Report
}

// probeInput анализирует видео и строит
// сетку ячеек.
func (task *job) probeInput() uint8 {
	info, detail, resCode := probe.Probe(
		task.ctx, task.args.PathInput,
	)
	task.info = info
	task.result.Detail = detail
	task.result.HasAudio = info.HasAudio
	task.grid = cell.NewGrid(info.Width, info.Height)
	isEmpty := resCode == rescode.OK &&
		task.grid.CntCells() == 0
	if isEmpty {
		resCode = rescode.ERR_PROBE_SIZE
	}
	return resCode
}

// cntLimitOf оценивает, сколько битов
// вмещает видео. Если число кадров
// неизвестно, ограничения нет.
func (task *job) cntLimitOf() int {
	cntLimit := math.MaxInt
	if task.info.CntFrames > 0 {
		cntLimit = (task.info.CntFrames + 1) *
			task.grid.CntCells()
	}
	return cntLimit
}

// recognize распознаёт биты ячеек кадра.
func (task *job) recognize(frame *ffpipe.Frame) {
	frame.Bits = cell.RecognizeFrame(
		frame.Pix, task.grid,
	)
}

// collect читает видео покадрово и копит
// поток скрытых данных.
func (task *job) collect() uint8 {
	var group sync.WaitGroup
	var detail string
	ctx, cancel := context.WithCancel(task.ctx)
	defer cancel()
	codeRead := rescode.OK
	cntCells := task.grid.CntCells()
	cntWorkers := ffpipe.CntWorkers()
	lane := ffpipe.NewLane(
		cntWorkers, task.grid.SizeLuma(),
	)
	pass := ffpipe.Pass{
		Args: ffpipe.ArgsGraySource(
			task.args.PathInput,
		),
		SizeFrame: task.grid.SizeLuma(),
	}
	stages := []func(){
		func() {
			_, detail, codeRead = ffpipe.ReadPass(
				ctx, pass, lane,
			)
			close(lane.Frames)
		},
		func() {
			ffpipe.RunWorkers(
				ctx, cntWorkers, lane, task.recognize,
			)
		},
	}
	for _, stage := range stages {
		group.Add(1)
		go func() {
			defer group.Done()
			stage()
		}()
	}
	taker := newCollector(
		task.args.Key, task.cntLimitOf(),
	)
	ffpipe.OrderFrames(
		lane,
		func(frame ffpipe.Frame) {
			isWanted := taker.isActive()
			taker.take(frame.Bits)
			if isWanted && taker.resCode == rescode.OK {
				task.onProgress(
					frame.Ind+1,
					(taker.cntNeed+cntCells-1)/cntCells,
				)
			}
			if !taker.isActive() {
				cancel()
			}
		},
	)
	group.Wait()
	task.stream = taker.stream
	task.result.Detail = detail
	resCode := taker.resCode
	if resCode == rescode.OK && !taker.isDone {
		resCode = rescode.ERR_STREAM_SHORT
	}
	if codeRead != rescode.OK && !taker.isDone {
		resCode = codeRead
	}
	return resCode
}

// parse восстанавливает данные из потока.
func (task *job) parse() uint8 {
	parsed, resCode := payload.Parse(
		task.stream, task.args.Key,
	)
	content := parsed.Content
	task.data = content.Data
	task.result.NameData = content.Name
	task.result.NameMode = parsed.Params.Name
	task.result.CntCopies = parsed.Params.CntCopies
	task.result.CntBytesData = len(content.Data)
	task.result.CntBytesStream = len(task.stream)
	task.result.CntBitsFixed =
		parsed.Stats.CntBitsFixed
	task.result.CntBitsBroken =
		parsed.Stats.CntBitsBroken
	task.result.CntBitsBch = content.CntBitsFixed
	task.result.CntBlocksFailed =
		content.CntBlocksFailed
	return resCode
}

// save записывает извлечённые данные.
func (task *job) save() uint8 {
	resCode := rescode.OK
	errWrite := os.WriteFile(
		task.args.PathOutput,
		task.data,
		MODE_FILE_OUTPUT,
	)
	if errWrite != nil {
		resCode = rescode.ERR_FILE_WRITE
	}
	return resCode
}

// Run выполняет команду split.
//
// Параметры:
//   - ctx: контекст отмены (прерывание
//     пользователем).
//   - args: аргументы команды; ключ
//     проверен, пути не пусты.
//   - onProgress: извещение о числе
//     разобранных кадров и числе нужных.
//
// Возвращает: отчёт с кодом результата:
// OK, ERR_PROC_*, ERR_PROBE_*, ERR_STREAM_*,
// ERR_DATA_CRC, ERR_FILE_WRITE.
func Run(
	ctx context.Context,
	args Args,
	onProgress func(cntDone, cntTotal int),
) report.Report {
	task := &job{
		ctx:        ctx,
		args:       args,
		onProgress: onProgress,
	}
	task.result.Code = rescode.Chain(
		ffpipe.CheckTools,
		task.probeInput,
		task.collect,
		task.parse,
		task.save,
	)
	isFailed := task.result.Code != rescode.OK
	if isFailed && ctx.Err() != nil {
		task.result.Code = rescode.ERR_PROC_STOP
	}
	return task.result
}
