package cell

// levelSet — множество уровней яркости.
type levelSet [CNT_WORDS_SET]uint64

// traits — признаки одной ячейки.
// setYes, setNo — уровни светлых пикселей
// каждого узора; sumYes, sumNo — разность
// сумм яркости светлых и тёмных пикселей;
// cntYes, cntNo — число тёмных пикселей,
// чей уровень встретился среди светлых.
type traits struct {
	setYes levelSet
	setNo  levelSet
	sumYes int
	sumNo  int
	cntYes int
	cntNo  int
}

// add добавляет уровень во множество.
func (set *levelSet) add(level uint8) {
	set[level>>CNT_SHIFT_WORD] |=
		1 << (level & MASK_WORD_INDEX)
}

// cntHits сообщает, сколько раз уровень
// входит во множество: 0 или 1.
func (set *levelSet) cntHits(level uint8) int {
	word := set[level>>CNT_SHIFT_WORD]
	return int(word >> (level & MASK_WORD_INDEX) & 1)
}

// forEachPixel обходит пиксели ячейки.
//
// Параметры:
//   - pix: плоскость яркости кадра.
//   - grid: сетка ячеек.
//   - indCell: номер ячейки.
//   - visit: обработчик пикселя: строка,
//     столбец и уровень яркости.
func forEachPixel(
	pix []byte,
	grid Grid,
	indCell int,
	visit func(rowPix, colPix int, level uint8),
) {
	rowTop := indCell / grid.CntCols * SIZE_CELL
	colLeft := indCell % grid.CntCols * SIZE_CELL
	for indPix := 0; indPix < SIZE_CELL_PIXELS; indPix++ {
		rowPix := rowTop + indPix/SIZE_CELL
		colPix := colLeft + indPix%SIZE_CELL
		visit(
			rowPix,
			colPix,
			pix[rowPix*grid.Width+colPix],
		)
	}
}

// signOf возвращает +1 для светлого пикселя
// узора и -1 для тёмного.
func signOf(isOn bool) int {
	sign := -1
	if isOn {
		sign = 1
	}
	return sign
}

// collect копит уровни светлых пикселей и
// разности сумм яркости для обоих узоров.
func (cell *traits) collect(
	rowPix, colPix int,
	level uint8,
) {
	isYes := IsOnYes(rowPix, colPix)
	isNo := IsOnNo(rowPix, colPix)
	cell.sumYes += signOf(isYes) * int(level)
	cell.sumNo += signOf(isNo) * int(level)
	if isYes {
		cell.setYes.add(level)
	}
	if isNo {
		cell.setNo.add(level)
	}
}

// compare считает тёмные пиксели, уровень
// которых встретился среди светлых.
func (cell *traits) compare(
	rowPix, colPix int,
	level uint8,
) {
	if !IsOnYes(rowPix, colPix) {
		cell.cntYes += cell.setYes.cntHits(level)
	}
	if !IsOnNo(rowPix, colPix) {
		cell.cntNo += cell.setNo.cntHits(level)
	}
}

// decide выносит решение о бите ячейки.
// Основное правило: узор принят, если ни
// один его тёмный пиксель не совпал по
// уровню со светлым. Если приняты оба узора
// или ни один, решает энергия полос.
//
// Возвращает: бит ячейки, 0 или 1.
func (cell *traits) decide() uint8 {
	isYes := cell.cntYes == 0
	isNo := cell.cntNo == 0
	if isYes == isNo {
		isYes = cell.sumYes > cell.sumNo
	}
	bit := uint8(0)
	if isYes {
		bit = 1
	}
	return bit
}

// RecognizeCell распознаёт бит одной ячейки.
//
// Параметры:
//   - pix: плоскость яркости кадра,
//     grid.SizeLuma() байтов.
//   - grid: сетка ячеек.
//   - indCell: номер ячейки,
//     0..grid.CntCells()-1.
//
// Возвращает: бит ячейки, 0 или 1.
func RecognizeCell(
	pix []byte,
	grid Grid,
	indCell int,
) uint8 {
	var cell traits
	forEachPixel(pix, grid, indCell, cell.collect)
	forEachPixel(pix, grid, indCell, cell.compare)
	return cell.decide()
}

// RecognizeFrame распознаёт биты всех ячеек.
//
// Параметры:
//   - pix: плоскость яркости кадра,
//     grid.SizeLuma() байтов.
//   - grid: сетка ячеек.
//
// Возвращает: по одному биту на ячейку,
// grid.CntCells() значений 0 или 1.
func RecognizeFrame(pix []byte, grid Grid) []uint8 {
	bits := make([]uint8, grid.CntCells())
	for indCell := range bits {
		bits[indCell] = RecognizeCell(
			pix, grid, indCell,
		)
	}
	return bits
}
