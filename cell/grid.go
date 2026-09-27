//============================================//
// Пакет cell                                 //
//                                            //
// Назначение:                                //
//   ячейки 8x8 пикселей, несущие по одному   //
//   биту. Единичный бит — горизонтальные     //
//   полосы, нулевой — вертикальные.          //
//   Рисование узоров на кадре yuv420p и      //
//   распознавание битов по яркости кадра.    //
//                                            //
// Автор: Aleksiej Ostrowski                  //
// Версия: 0.1.0                              //
//============================================//

package cell

// Grid — сетка ячеек на кадре.
// Width, Height — размер кадра в пикселях;
// CntCols, CntRows — число полных ячеек.
type Grid struct {
	Width   int
	Height  int
	CntCols int
	CntRows int
}

// PATTERNS — узоры по значению бита.
var PATTERNS = [CNT_PATTERNS]func(int, int) bool{
	IsOnNo,
	IsOnYes,
}

// NewGrid строит сетку ячеек для кадра.
//
// Параметры:
//   - width: ширина кадра, пиксели, >= 0.
//   - height: высота кадра, пиксели, >= 0.
//
// Возвращает: сетку полных ячеек.
func NewGrid(width, height int) Grid {
	return Grid{
		Width:   width,
		Height:  height,
		CntCols: width / SIZE_CELL,
		CntRows: height / SIZE_CELL,
	}
}

// CntCells считает полные ячейки кадра.
//
// Возвращает: число ячеек, оно же число
// битов в одном кадре.
func (grid Grid) CntCells() int {
	return grid.CntCols * grid.CntRows
}

// SizeLuma считает байты плоскости яркости.
//
// Возвращает: размер плоскости в байтах.
func (grid Grid) SizeLuma() int {
	return grid.Width * grid.Height
}

// SizeFrame считает байты кадра yuv420p
// с чётными шириной и высотой.
//
// Возвращает: размер кадра в байтах.
func (grid Grid) SizeFrame() int {
	sizeChroma := (grid.Width >> CNT_SHIFT_CHROMA) *
		(grid.Height >> CNT_SHIFT_CHROMA)
	return grid.SizeLuma() +
		CNT_PLANES_CHROMA*sizeChroma
}

// IsOnNo — узор нулевого бита: светлые
// пиксели образуют вертикальные полосы.
//
// Параметры:
//   - rowPix: строка пикселя, >= 0.
//   - colPix: столбец пикселя, >= 0.
//
// Возвращает: true для светлого пикселя.
func IsOnNo(rowPix, colPix int) bool {
	return colPix&1 == 0
}

// IsOnYes — узор единичного бита: светлые
// пиксели образуют горизонтальные полосы.
//
// Параметры:
//   - rowPix: строка пикселя, >= 0.
//   - colPix: столбец пикселя, >= 0.
//
// Возвращает: true для светлого пикселя.
func IsOnYes(rowPix, colPix int) bool {
	return rowPix&1 == 0
}
