/*

# ------------------------------ #
#                                #
#  version 0.0.1                 #
#                                #
#  Aleksiej Ostrowski, 2023      #
#                                #
#  https://aleksiej.com          #
#                                #
# ------------------------------ #

*/

package recognize

import (
	// "fmt"
	"image"
	"image/color"
	// "math/rand"
	// "math"
	// "github.com/lucasb-eyer/go-colorful"
)

var YES = [64]uint8{
	255, 255, 255, 255, 255, 255, 255, 255,
	0, 0, 0, 0, 0, 0, 0, 0,
	255, 255, 255, 255, 255, 255, 255, 255,
	0, 0, 0, 0, 0, 0, 0, 0,
	255, 255, 255, 255, 255, 255, 255, 255,
	0, 0, 0, 0, 0, 0, 0, 0,
	255, 255, 255, 255, 255, 255, 255, 255,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var NO = [64]uint8{
	255, 0, 255, 0, 255, 0, 255, 0,
	255, 0, 255, 0, 255, 0, 255, 0,
	255, 0, 255, 0, 255, 0, 255, 0,
	255, 0, 255, 0, 255, 0, 255, 0,
	255, 0, 255, 0, 255, 0, 255, 0,
	255, 0, 255, 0, 255, 0, 255, 0,
	255, 0, 255, 0, 255, 0, 255, 0,
	255, 0, 255, 0, 255, 0, 255, 0,
}

func Recognize8(img *image.Image, X, Y int) int {

	YES_colors := make(map[color.RGBA]bool)
	NO_colors := make(map[color.RGBA]bool)

	for y := 0; y < 8; y++ {
		sh := y << 3
		for x := 0; x < 8; x++ {
			idx := sh + x
			flag1 := YES[idx] == 255
			flag2 := NO[idx] == 255
			cl := color.RGBAModel.Convert((*img).At(X+x, Y+y)).(color.RGBA)
			if flag1 {
				YES_colors[cl] = true
			}
			if flag2 {
				NO_colors[cl] = true
			}
		}
	}

	coef_YES_MINUS := 0
	coef_NO_MINUS := 0

	for y := 0; y < 8; y++ {
		sh := y << 3
		for x := 0; x < 8; x++ {
			idx := sh + x
			flag1 := YES[idx] == 255
			flag2 := NO[idx] == 255
			cl := color.RGBAModel.Convert((*img).At(X+x, Y+y)).(color.RGBA)

			if !flag1 {
				if _, ok := YES_colors[cl]; ok {
					coef_YES_MINUS++
				}
			}

			if !flag2 {
				if _, ok := NO_colors[cl]; ok {
					coef_NO_MINUS++
				}
			}
		}
	}

	res := 0
	flag := 0
	if coef_YES_MINUS == 0 {
		res = 1
		flag++
	}
	if coef_NO_MINUS == 0 {
		res = -1
		flag++
	}
	if flag == 2 {
		res = 0
	}

	// fmt.Println(coef_YES_MINUS, coef_NO_MINUS)

	return res
}
