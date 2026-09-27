package wrap

// splitNibbles раскладывает байты на
// полубайты, старший полубайт первым.
//
// Параметры:
//   - src: исходные байты.
//
// Возвращает: полубайты, вдвое длиннее src.
func splitNibbles(src []byte) []uint8 {
	nibbles := make(
		[]uint8,
		len(src)*CNT_NIBBLES_BYTE,
	)
	for indSrc, value := range src {
		indDst := indSrc * CNT_NIBBLES_BYTE
		nibbles[indDst] = value >> CNT_BITS_NIBBLE
		nibbles[indDst+1] = value & MASK_NIBBLE
	}
	return nibbles
}

// joinNibbles собирает байты из пар
// полубайтов, старший полубайт первым.
//
// Параметры:
//   - nibbles: полубайты, чётное количество.
//
// Возвращает: байты, вдвое короче nibbles.
func joinNibbles(nibbles []uint8) []byte {
	dst := make(
		[]byte,
		len(nibbles)/CNT_NIBBLES_BYTE,
	)
	for indDst := range dst {
		indSrc := indDst * CNT_NIBBLES_BYTE
		dst[indDst] =
			nibbles[indSrc]<<CNT_BITS_NIBBLE |
				nibbles[indSrc+1]
	}
	return dst
}

// walkOrder обходит ячейки перемежения:
// шаг 1980 по кругу, занятая ячейка
// пропускается линейным поиском.
//
// Параметры:
//   - cntNibbles: число полубайтов, >= 0.
//   - visit: обработчик очередной ячейки:
//     номер шага обхода и номер ячейки.
func walkOrder(
	cntNibbles int,
	visit func(indOrder, slot int),
) {
	isTaken := make([]bool, cntNibbles)
	pos := CNT_STEP_NIBBLE
	for indOrder := range isTaken {
		for isTaken[pos%cntNibbles] {
			pos++
		}
		slot := pos % cntNibbles
		isTaken[slot] = true
		visit(indOrder, slot)
		pos = (pos + CNT_STEP_NIBBLE) % cntNibbles
	}
}

// interleave перемежает полубайты: исходные
// берутся с конца и раскладываются по
// ячейкам в порядке обхода.
//
// Параметры:
//   - src: исходные полубайты.
//
// Возвращает: перемежённые полубайты.
func interleave(src []uint8) []uint8 {
	dst := make([]uint8, len(src))
	indLast := len(src) - 1
	walkOrder(len(src), func(indOrder, slot int) {
		dst[slot] = src[indLast-indOrder]
	})
	return dst
}

// deinterleave выполняет обратное
// перемежение полубайтов.
//
// Параметры:
//   - src: перемежённые полубайты.
//
// Возвращает: полубайты в исходном порядке.
func deinterleave(src []uint8) []uint8 {
	dst := make([]uint8, len(src))
	indLast := len(src) - 1
	walkOrder(len(src), func(indOrder, slot int) {
		dst[indLast-indOrder] = src[slot]
	})
	return dst
}
