//============================================//
// Пакет report                               //
//                                            //
// Назначение:                                //
//   отчёт о выполнении команды. Функции      //
//   программы ничего не печатают: они        //
//   возвращают отчёт, который выводится      //
//   в единственном месте — в main.           //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package report

// Report — отчёт о выполнении команды.
// Code — код результата (см. rescode);
// Detail — пояснение внешней программы;
// прочие поля — итоговые показатели.
type Report struct {
	Code            uint8
	Detail          string
	NameData        string
	NameMode        string
	CntBytesData    int
	CntBytesStream  int
	CntCopies       int
	CntFrames       int
	CntPasses       int
	HasAudio        bool
	CntBitsFixed    int
	CntBitsBroken   int
	CntBitsBch      int
	CntBlocksFailed int
}
