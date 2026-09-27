//============================================//
// Пакет repeat                               //
//                                            //
// Назначение:                                //
//   код с N-кратным повторением: данные      //
//   записываются N копиями подряд, при       //
//   восстановлении каждый бит определяется   //
//   голосованием по всем копиям.             //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package repeat

import "math"

// Stats — статистика восстановления.
// CntBitsFixed — биты, исправленные
// уверенным большинством; CntBitsBroken —
// биты без уверенного большинства
// (пессимистическая оценка потерь).
type Stats struct {
	CntBitsFixed  int
	CntBitsBroken int
}

// verdict — решение для одного числа
// единичных голосов.
type verdict struct {
	bit      uint8
	cntFixed int
	cntBroke int
}

// Repeat записывает данные N копиями подряд.
//
// Параметры:
//   - src: исходные данные.
//   - cntCopies: число копий, >= 1.
//
// Возвращает: поток длиной len(src) * N.
func Repeat(src []byte, cntCopies int) []byte {
	stream := make([]byte, 0, len(src)*cntCopies)
	for indCopy := 0; indCopy < cntCopies; indCopy++ {
		stream = append(stream, src...)
	}
	return stream
}

// cntAcceptOf вычисляет число согласных
// голосов, достаточное для уверенности.
//
// Параметры:
//   - cntCopies: число копий, >= 1.
//
// Возвращает: порог уверенности в голосах.
func cntAcceptOf(cntCopies int) int {
	cntAccept := cntCopies
	if cntCopies != CNT_COPIES_PAIR {
		share := float64(cntCopies) *
			LIM_SHARE_ACCEPT
		cntAccept = int(math.Round(share))
	}
	return cntAccept
}

// judge выносит решение по числу единичных
// голосов: бит — по простому большинству,
// пометки — по порогу уверенности.
//
// Параметры:
//   - cntOnes: число единиц, 0..cntCopies.
//   - cntCopies: число копий, >= 1.
//
// Возвращает: решение для бита.
func judge(cntOnes, cntCopies int) verdict {
	var result verdict
	cntAccept := cntAcceptOf(cntCopies)
	cntMajor := max(cntOnes, cntCopies-cntOnes)
	isSplit := cntMajor != cntCopies
	isSure := cntMajor >= cntAccept
	if cntOnes<<1 > cntCopies {
		result.bit = 1
	}
	if isSplit && isSure {
		result.cntFixed = 1
	}
	if isSplit && !isSure {
		result.cntBroke = 1
	}
	return result
}

// Restore восстанавливает данные голосованием
// по каждому биту.
//
// Параметры:
//   - stream: N копий подряд; длина
//     кратна cntCopies.
//   - cntCopies: число копий, >= 1.
//
// Возвращает: данные длиной len(stream) / N
// и статистику восстановления.
func Restore(
	stream []byte,
	cntCopies int,
) ([]byte, Stats) {
	var stats Stats
	verdicts := make([]verdict, cntCopies+1)
	for cntOnes := range verdicts {
		verdicts[cntOnes] = judge(cntOnes, cntCopies)
	}
	sizeCopy := len(stream) / cntCopies
	restored := make([]byte, sizeCopy)
	for indByte := range restored {
		var votes [CNT_BITS_BYTE]int
		for indCopy := 0; indCopy < cntCopies; indCopy++ {
			value := stream[indCopy*sizeCopy+indByte]
			for indBit := range votes {
				votes[indBit] +=
					int(value>>indBit) & 1
			}
		}
		for indBit, cntOnes := range votes {
			result := verdicts[cntOnes]
			restored[indByte] |= result.bit << indBit
			stats.CntBitsFixed += result.cntFixed
			stats.CntBitsBroken += result.cntBroke
		}
	}
	return restored, stats
}
