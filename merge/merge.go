//============================================//
// Пакет merge                                //
//                                            //
// Назначение:                                //
//   команда merge — сокрытие файла данных    //
//   в видео со звуковой дорожкой.            //
//                                            //
// Входные данные:                            //
//   ключ, режим, файл данных, исходное видео //
// Выходные данные:                           //
//   видео H.264 + AAC со скрытыми данными    //
//                                            //
// Порядок работы:                            //
//   1. данные -> БЧХ -> копии -> перемеши-   //
//      вание -> поток с заголовком;          //
//   2. декодер отдаёт кадры по одному,       //
//      потребители рисуют ячейки, кадры      //
//      уходят кодировщику; на диск кадры     //
//      не распаковываются;                   //
//   3. исходное видео повторяется целое      //
//      число раз; звук каждого повтора       //
//      подгоняется под длительность видео.   //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package merge

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"stegano2/bitio"
	"stegano2/cell"
	"stegano2/ffpipe"
	"stegano2/mode"
	"stegano2/payload"
	"stegano2/probe"
	"stegano2/report"
	"stegano2/rescode"
	"stegano2/stir"
)

// Args — аргументы команды merge.
type Args struct {
	Key          string
	Params       mode.Params
	PathData     string
	PathOriginal string
	PathResult   string
}

// verdict — первая ошибка конвейера.
type verdict struct {
	mutex  sync.Mutex
	code   uint8
	detail string
	cancel context.CancelFunc
}

// job — состояние выполняемой команды.
type job struct {
	ctx        context.Context
	args       Args
	onProgress func(cntDone, cntTotal int)
	stream     []byte
	info       probe.Info
	plan       Plan
	palette    cell.Palette
	pathTemp   string
	result     report.Report
}

// fail запоминает первую ошибку и
// останавливает конвейер.
func (first *verdict) fail(
	code uint8,
	detail string,
) {
	first.mutex.Lock()
	if first.code == rescode.OK {
		first.code, first.detail = code, detail
	}
	first.mutex.Unlock()
	first.cancel()
}

// check передаёт ошибку этапа, если она есть.
func (first *verdict) check(
	code uint8,
	detail string,
) {
	if code != rescode.OK {
		first.fail(code, detail)
	}
}

// loadStream читает данные и строит поток.
func (task *job) loadStream() uint8 {
	resCode := rescode.OK
	data, errRead := os.ReadFile(task.args.PathData)
	if errRead != nil {
		resCode = rescode.ERR_FILE_READ
	}
	if resCode == rescode.OK {
		name := filepath.Base(task.args.PathData)
		task.stream, resCode = payload.Build(
			name,
			data,
			task.args.Key,
			task.args.Params,
		)
		task.result.NameData = name
		task.result.CntBytesData = len(data)
		task.result.CntBytesStream = len(task.stream)
	}
	return resCode
}

// probeOriginal анализирует исходное видео.
func (task *job) probeOriginal() uint8 {
	info, detail, resCode := probe.Probe(
		task.ctx, task.args.PathOriginal,
	)
	task.info = info
	task.result.Detail = detail
	task.result.HasAudio = info.HasAudio
	return resCode
}

// makePlan строит план и таблицы смешивания.
func (task *job) makePlan() uint8 {
	plan, resCode := MakePlan(
		task.info, len(task.stream),
	)
	task.plan = plan
	task.palette = cell.NewPalette(
		task.args.Params.Mix,
	)
	task.pathTemp = filepath.Join(
		filepath.Dir(task.args.PathResult),
		PREFIX_TEMP+
			filepath.Base(task.args.PathResult),
	)
	task.result.CntFrames = plan.CntFrames
	task.result.CntPasses = plan.CntPasses
	return resCode
}

// drawFrame рисует на кадре ячейки с битами
// потока; свободные ячейки получают
// псевдослучайные биты.
func (task *job) drawFrame(frame *ffpipe.Frame) {
	grid := task.plan.Grid
	indFirst := frame.Ind * grid.CntCells()
	cntBits := bitio.CntBits(len(task.stream))
	seed := uint64(stir.SeedOf(task.args.Key))
	cell.DrawFrame(
		frame.Pix,
		grid,
		&task.palette,
		func(indCell int) uint8 {
			indBit := indFirst + indCell
			bit := DustBit(seed, indBit)
			if indBit < cntBits {
				bit = bitio.Get(task.stream, indBit)
			}
			return bit
		},
	)
}

// readVideo отправляет в конвейер кадры всех
// проходов исходного видео и закрывает
// канал кадров.
func (task *job) readVideo(
	ctx context.Context,
	lane ffpipe.Lane,
	first *verdict,
) {
	pass := ffpipe.Pass{
		Args: ffpipe.ArgsVideoSource(
			task.args.PathOriginal, task.info,
		),
		SizeFrame: task.plan.Grid.SizeFrame(),
		CntFrames: task.plan.CntFramesPass,
	}
	indPass := 0
	for ctx.Err() == nil && indPass < task.plan.CntPasses {
		pass.IndFirst = indPass * pass.CntFrames
		_, detail, resCode := ffpipe.ReadPass(
			ctx, pass, lane,
		)
		first.check(resCode, detail)
		indPass++
	}
	close(lane.Frames)
}

// pumpAudio передаёт кодировщику звук всех
// проходов и закрывает канал звука.
func (task *job) pumpAudio(
	ctx context.Context,
	encoder *ffpipe.Encoder,
	first *verdict,
) {
	args := ffpipe.ArgsAudioSource(
		task.args.PathOriginal, task.info,
	)
	indPass := 0
	for ctx.Err() == nil && indPass < task.plan.CntPasses {
		detail, resCode := ffpipe.PumpAudio(
			ctx,
			args,
			SizeAudioPass(
				task.info, task.plan, indPass,
			),
			encoder.Audio,
		)
		first.check(resCode, detail)
		indPass++
	}
	encoder.Audio.Close()
}

// writeVideo передаёт кодировщику кадры
// в исходном порядке и закрывает канал
// кадров кодировщика.
func (task *job) writeVideo(
	ctx context.Context,
	lane ffpipe.Lane,
	encoder *ffpipe.Encoder,
	first *verdict,
) {
	cntDone := 0
	ffpipe.OrderFrames(
		lane,
		func(frame ffpipe.Frame) {
			var errWrite error
			isAlive := ctx.Err() == nil
			if isAlive {
				_, errWrite = encoder.Video.Write(
					frame.Pix,
				)
			}
			if errWrite != nil {
				first.fail(
					rescode.ERR_PIPE_WRITE,
					errWrite.Error(),
				)
			}
			if isAlive && errWrite == nil {
				cntDone++
				task.onProgress(
					cntDone, task.plan.CntFrames,
				)
			}
		},
	)
	encoder.Video.Close()
}

// encode строит видео: запускает кодировщик
// и конвейер кадров, ждёт их завершения.
func (task *job) encode() uint8 {
	ctx, cancel := context.WithCancel(task.ctx)
	defer cancel()
	first := &verdict{cancel: cancel}
	grid := task.plan.Grid
	encoder, resCode := ffpipe.StartEncoder(
		ctx,
		ffpipe.ArgsEncoder(
			task.pathTemp,
			grid.Width,
			grid.Height,
			task.info,
			task.args.Params.Crf,
		),
		task.info.HasAudio,
	)
	if resCode == rescode.OK {
		var group sync.WaitGroup
		cntWorkers := ffpipe.CntWorkers()
		lane := ffpipe.NewLane(
			cntWorkers, grid.SizeFrame(),
		)
		stages := []func(){
			func() {
				task.readVideo(ctx, lane, first)
			},
			func() {
				ffpipe.RunWorkers(
					ctx,
					cntWorkers,
					lane,
					task.drawFrame,
				)
			},
			func() {
				task.writeVideo(
					ctx, lane, encoder, first,
				)
			},
		}
		if task.info.HasAudio {
			stages = append(stages, func() {
				task.pumpAudio(ctx, encoder, first)
			})
		}
		for _, stage := range stages {
			group.Add(1)
			go func() {
				defer group.Done()
				stage()
			}()
		}
		group.Wait()
		detail, codeEncoder := encoder.Finish()
		resCode = first.code
		task.result.Detail = first.detail
		isEncoder := resCode == rescode.OK ||
			resCode == rescode.ERR_PIPE_WRITE
		if isEncoder && codeEncoder != rescode.OK {
			resCode = codeEncoder
			task.result.Detail = detail
		}
	}
	return resCode
}

// publish переименовывает готовое видео.
func (task *job) publish() uint8 {
	resCode := rescode.OK
	errRename := os.Rename(
		task.pathTemp, task.args.PathResult,
	)
	if errRename != nil {
		resCode = rescode.ERR_FILE_RENAME
	}
	return resCode
}

// Run выполняет команду merge.
//
// Параметры:
//   - ctx: контекст отмены (прерывание
//     пользователем).
//   - args: аргументы команды; ключ и режим
//     проверены, пути не пусты.
//   - onProgress: извещение о числе
//     готовых кадров и их общем числе.
//
// Возвращает: отчёт с кодом результата:
// OK, ERR_PROC_*, ERR_PIPE_*, ERR_FILE_*,
// ERR_DATA_*, ERR_PROBE_*.
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
	task.result.NameMode = args.Params.Name
	task.result.CntCopies = args.Params.CntCopies
	task.result.Code = rescode.Chain(
		ffpipe.CheckTools,
		task.loadStream,
		task.probeOriginal,
		task.makePlan,
		task.encode,
		task.publish,
	)
	isFailed := task.result.Code != rescode.OK
	if isFailed {
		os.Remove(task.pathTemp)
	}
	if isFailed && ctx.Err() != nil {
		task.result.Code = rescode.ERR_PROC_STOP
	}
	return task.result
}
