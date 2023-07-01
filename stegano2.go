/*

# ------------------------------ #
#                                #
#  version 0.0.2                 #
#                                #
#  Aleksiej Ostrowski, 2023      #
#                                #
#  https://aleksiej.com          #
#                                #
# ------------------------------ #

*/

package main

import (
	copyXN "./copyXN"
	imaging "./imaging"
	recognize "./recognize"
	stir "./stir"
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"image"
	// "image/color"
	"image/draw"
	// "image/png"
	"golang.org/x/image/bmp"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BUF_SIZE = 4096
)

type TMainParams struct {
	info                   string
	logFile                string
	tempDir                string
	tempDir_audio          string
	tempDir_rpics          string
	tempDir_bpics          string
	tempDir_res            string
	max_rpics              uint32
	max_bpics              uint32
	width_pics             int
	height_pics            int
	square_size            int
	back_filename          string
	duration_back_filename uint32
	binary_input           string
	binary_output          string
	video_input            string
	video_output           string
	xN_input               string
	xN_input_stir          string
	xN_output              string
	xN_output_stir         string
	N                      int
	broken_bits            int64
	successful_corrections int64
	encode_decode          int // 1=encode, 2=decode
	mode                   int // 1=aggressive(for YouTube experiments), 2=experimental, 3=comfortable(only for the current file save)
	code_key               string
	mix                    float64
	shuffle_bits           bool
}

func saveImg(img image.Image, pic_num uint32, prms TMainParams) error {

	fn := filepath.Join(prms.tempDir_rpics, fmt.Sprintf("pic_%0*d.bmp", 8, pic_num))
	file, err := os.Create(fn)

	if err != nil {
		return fmt.Errorf("Сannot create a file %s", fn)
	}

	defer file.Close()
	err = bmp.Encode(file, img)
	if err != nil {
		return fmt.Errorf("Сannot encode a file %s", fn)
	}

	return nil
}

func GetPNGDimensions(filename string) (int, int, error) {

	file, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return 0, 0, fmt.Errorf("Сannot open a file %s", filename)
	}
	defer file.Close()

	imageData, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("Сannot decode a file %s", filename)
	}

	return imageData.Width, imageData.Height, nil
}

/*

func convertGrayToNRGBA(grayImg *image.Gray) *image.NRGBA {

	bounds := grayImg.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	nrgbaImg := image.NewNRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gV := grayImg.GrayAt(x, y).Y
			nrgbaImg.SetNRGBA(x, y, color.NRGBA{gV, gV, gV, 255})
		}
	}

	return nrgbaImg
}
*/

func convertGrayToNRGBA(src *image.Gray) *image.NRGBA {
	dst := image.NewNRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	return dst
}

func makeMix(img_ *[]uint8, pic_num uint32, prms TMainParams) (image.Image, error) {

	img := image.NewGray(image.Rect(0, 0, prms.width_pics, prms.height_pics))
	copy(img.Pix, *img_)

	back, err := selectBack(pic_num, prms)
	if err != nil {
		return nil, err
	}

	return imaging.Overlay(back, convertGrayToNRGBA(img), image.Pt(0, 0), prms.mix), nil
}

func countFiles(dir string) (uint32, error) {
	fileCount := uint32(0)

	matches, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return 0, err
	}

	fileCount = uint32(len(matches))

	return fileCount, nil
}

func selectBack(pic_num uint32, prms TMainParams) (image.Image, error) {

	if pic_num >= prms.max_bpics {
		pic_num = pic_num % prms.max_bpics
		if pic_num == 0 {
			pic_num = prms.max_bpics
		}
	}

	back, err := os.OpenFile(filepath.Join(prms.tempDir_bpics, fmt.Sprintf("pic_%0*d.bmp", 8, pic_num)), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer back.Close()

	img_back, err := bmp.Decode(back)
	if err != nil {
		return nil, err
	}

	return img_back, nil
}

type saveImgStruct struct {
	img image.Image
	num uint32
}

func bin2pics(prms TMainParams) (int64, error) {

	inp, err := os.OpenFile(prms.xN_input_stir, os.O_RDONLY, 0600)
	if err != nil {
		return int64(0), err
	}

	defer inp.Close()

	st, err := inp.Stat()
	if err != nil {
		return int64(0), err
	}

	if st.Size() == 0 {
		return int64(0), fmt.Errorf("Binary file %s is empty", prms.xN_input_stir)
	}

	manager := make(chan int64)
	defer close(manager)

	reader := bufio.NewReader(inp)

	bit_chan := make(chan uint8)
	go func(reader *bufio.Reader, bit_chan chan<- uint8, prms TMainParams) {

		buf := make([]uint8, BUF_SIZE)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				for _, item := range buf[:n] {
					for bit := 0; bit < 8; bit++ {
						bit_chan <- (item >> bit) & 1
					}
				}
			}
			if err != nil {
				break
			}
		}

		close(bit_chan)

	}(reader, bit_chan, prms)

	go func(bit_chan <-chan uint8, manager chan<- int64, prms TMainParams) {

		img_ := make([]uint8, prms.width_pics*prms.height_pics)

		how_w := prms.width_pics / prms.square_size
		how_hw := prms.height_pics * how_w / prms.square_size

		offset, idx_w, idx_hw := 0, 0, 0
		pic_num, fs := uint32(1), int64(0)

		hlp_cnst0 := prms.width_pics * (prms.square_size - 1)
		hlp_cnst1, hlp_cnst2 := make([]int, prms.square_size), make([]int, prms.square_size)

		for sq := 0; sq < prms.square_size; sq++ {
			hlp_cnst1[sq] = prms.width_pics * sq
			hlp_cnst2[sq] = prms.square_size * sq
		}

		make_pic := func(bit uint8, offset int, img_ *[]uint8, wg *sync.WaitGroup) {
			defer (*wg).Done()
			LABEL := recognize.NO
			if bit == 1 {
				LABEL = recognize.YES
			}
			for sq := 0; sq < prms.square_size; sq++ {
				hlp_cnst3 := offset + hlp_cnst1[sq]
				copy((*img_)[hlp_cnst3:hlp_cnst3+prms.square_size],
					LABEL[hlp_cnst2[sq]:hlp_cnst2[sq]+prms.square_size])
			}
		}

		save_chan := make(chan saveImgStruct)
		defer close(save_chan)

		go func(save_chan <-chan saveImgStruct) {
			for el := range save_chan {
				err := saveImg(el.img, el.num, prms)
				if err != nil {
					log.Error(err)
					panic(err)
				}
			}
		}(save_chan)

		save_pic := func(save_chan chan<- saveImgStruct) {
			offset, idx_w, idx_hw = 0, 0, 0
			img_mix, err := makeMix(&img_, pic_num, prms)
			if err != nil {
				log.Error(err)
				panic(err)
			}
			var el saveImgStruct

			el.img = img_mix
			el.num = pic_num

			save_chan <- el

			img_ = make([]uint8, prms.width_pics*prms.height_pics)
			pic_num += 1
		}

		var wg sync.WaitGroup

		add_offset := func() {
			offset += prms.square_size
			idx_w += 1
			if idx_w >= how_w {
				idx_w = 0
				offset += hlp_cnst0
			}
		}

		for bit := range bit_chan {
			fs++
			wg.Add(1)
			go make_pic(bit, offset, &img_, &wg)
			add_offset()
			idx_hw += 1

			/*
				if idx_hw%runtime.GOMAXPROCS(0) == 0 {
					wg.Wait()
				}
			*/

			if idx_hw < how_hw {
				continue
			}

			wg.Wait()
			save_pic(save_chan)
		}

		if idx_hw > 0 {
			for dust := idx_hw; dust < how_hw; dust++ {
				wg.Add(1)
				go make_pic(uint8(rand.Intn(2)), offset, &img_, &wg)
				add_offset()
			}

			wg.Wait()
			save_pic(save_chan)
		}
		manager <- fs

	}(bit_chan, manager, prms)

	fs := <-manager

	return fs, nil
}

func selectNumFromFileName(filename string) (int64, error) {
	re := regexp.MustCompile(`\[[^\[\]]*\]`)
	for re.MatchString(filename) {
		filename = re.ReplaceAllString(filename, "")
	}
	start := strings.LastIndex(filename, "filesize") // filesize in bits
	if start == -1 {
		err := fmt.Errorf("There is no 'filesizeXX..X' in file %s", filename)
		return 0, err
	}
	end := strings.LastIndex(filename, ".")
	if end == -1 || end < start {
		err := fmt.Errorf("There is no a symbol '.' after 'filesizeXX..X' in file %s", filename)
		return 0, err
	}
	pattern := strings.TrimSpace(filename[start+8 : end])
	// fmt.Println("filename[start+8:end] = ", pattern)
	res, err := strconv.ParseInt(pattern, 10, 64)
	if err != nil {
		err := fmt.Errorf("An error occurred while parsing 'filesizeXX..X', size 'XX..X' = %s", pattern)
		return 0, err
	}
	return res, nil
}

type byteStruct struct {
	rs  int
	num int
}

func decodeVideo(prms TMainParams) error {

	fs, err := selectNumFromFileName(prms.video_input)
	if err != nil {
		return err
	}

	file_result, err := os.OpenFile(prms.xN_output_stir, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file_result)
	defer func() {
		writer.Flush()
		file_result.Close()
	}()

	cnt := int64(0)
	pic_num := uint32(0)

	for {

		pic_num += 1

		file, err := os.OpenFile(filepath.Join(prms.tempDir_rpics, fmt.Sprintf("pic_%0*d.bmp", 8, pic_num)), os.O_RDONLY, 0600)
		if err != nil {
			return err
		}
		defer file.Close()

		img, err := bmp.Decode(file)
		if err != nil {
			return err
		}

		bounds := img.Bounds()
		width_, height_ := bounds.Max.X, bounds.Max.Y
		save_byte, num_bit := uint8(0), 0
		byte_chan := make(chan byteStruct, 8)

		var wg sync.WaitGroup

		for y := 0; y <= height_-prms.square_size; y += prms.square_size {
			for x := 0; x <= width_-prms.square_size; x += prms.square_size {
				wg.Add(1)
				go func(x, y, num_bit int) {
					defer wg.Done()
					var res byteStruct
					res.rs = recognize.Recognize8(&img, x, y)
					res.num = num_bit
					byte_chan <- res
				}(x, y, num_bit)

				num_bit++

				if num_bit <= 7 {
					continue
				}

				wg.Wait()

				for len(byte_chan) > 0 {
					el := <-byte_chan
					rs := el.rs
					if rs == 0 {
						rs = rand.Intn(2)
					}
					if rs == 1 {
						save_byte |= 1 << el.num
					}
				}

				err := writer.WriteByte(save_byte)
				if err != nil {
					return err
				}

				save_byte, num_bit = uint8(0), 0
				cnt += 8

				if cnt >= fs {
					return nil
				}
			}
		}

		if cnt >= fs {
			return nil
		}
	}

}

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	MainParams := TMainParams{}
	MainParams.info = `

#----------------------------#
#                            #
#  version 0.0.2             #
#                            #
#  Aleksiej Ostrowski, 2023  #
#                            #
#  https://aleksiej.com      #
#                            #
#----------------------------#

For encrypting a binary file:

./stegano2 e 1 UNICKEY binary_file_name back_video_file_name.webm

For decrypting a video file:

./stegano2 d 1 UNICKEY video_file_name.webm

`
	var (
		code int = -1
		err  error
	)

	defer func() {

		if r := recover(); r != nil {
			err = fmt.Errorf("It was panic")
			code = 3
		}

		switch code {
		case 0:
			if MainParams.encode_decode == 2 {
				fmt.Println("Totally broken bits (pessimistic assessment) = ", MainParams.broken_bits)
				fmt.Println("Successful corrections bits = ", MainParams.successful_corrections)
			}
			fmt.Println("ok")
		case 1:
			fmt.Println(MainParams.info)
		case 2:
			fmt.Println("The length of the UNIKEY key must be at least 3 characters")
		case -1, 3:
			log.Error(err)
			fmt.Errorf("%s", err)
		case 4:
			fmt.Println("1=aggressive(for YouTube experiments), 2=experimental, 3=comfortable(only for the current file save)")
		default:
			fmt.Printf("Error %d\n", code)
		}

		n := runtime.NumGoroutine()
		if n != 1 {
			fmt.Println("Warning! There are working routines, n = ", n)
		}

		os.RemoveAll(MainParams.tempDir)

		os.Exit(code)
	}()

	MainParams.logFile = "stegano2.log"

	{
		var file, err = os.OpenFile(MainParams.logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

		if err != nil {
			fmt.Println("log file was not opened", err)
		}

		defer file.Close()

		log.SetOutput(file)
		log.Print("Init log file...")
	}

	rand.Seed(time.Now().UnixNano())

	MainParams.tempDir, err = ioutil.TempDir("", "stegano_"+strconv.Itoa(rand.Intn(math.MaxInt)))
	if err != nil {
		return
	}

	MainParams.tempDir_audio = filepath.Join(MainParams.tempDir, "audio")
	if err = os.Mkdir(MainParams.tempDir_audio, 0700); err != nil {
		return
	}

	MainParams.tempDir_rpics = filepath.Join(MainParams.tempDir, "rpics")
	if err = os.Mkdir(MainParams.tempDir_rpics, 0700); err != nil {
		return
	}

	MainParams.tempDir_bpics = filepath.Join(MainParams.tempDir, "bpics")
	if err = os.Mkdir(MainParams.tempDir_bpics, 0700); err != nil {
		return
	}

	MainParams.tempDir_res = filepath.Join(MainParams.tempDir, "res")
	if err = os.Mkdir(MainParams.tempDir_res, 0700); err != nil {
		return
	}

	for _, necessary := range []string{"ffmpeg"} { // + "ffprobe"
		if _, err = exec.LookPath(necessary); err != nil {
			return
		}
	}

	len_osArgs := len(os.Args)

	if len_osArgs != 5 && len_osArgs != 6 {
		code = 1
		return
	}

	switch os.Args[1] {
	case "e":
		MainParams.encode_decode = 1
	case "d":
		MainParams.encode_decode = 2
	}

	switch MainParams.encode_decode {
	case 1:
		if len_osArgs != 6 {
			code = 1
			return
		}
		MainParams.binary_input = os.Args[4]
		MainParams.back_filename = os.Args[5]
		MainParams.xN_input = filepath.Join(MainParams.tempDir_res, "binary_input.XN")
		MainParams.xN_input_stir = filepath.Join(MainParams.tempDir_res, "binary_input.XN.stir")
	case 2:
		if len_osArgs != 5 {
			code = 1
			return
		}
		MainParams.video_input = os.Args[4]
		MainParams.binary_output = MainParams.video_input + ".original"
		MainParams.xN_output = filepath.Join(MainParams.tempDir_res, "binary_output.XN")
		MainParams.xN_output_stir = filepath.Join(MainParams.tempDir_res, "binary_output.XN.stir")
	default:
		code = 1
		return
	}

	MainParams.mode, err = strconv.Atoi(os.Args[2])
	if err != nil {
		code = 1
		return
	}

	MainParams.code_key = os.Args[3]
	if len(MainParams.code_key) <= 2 {
		code = 2
		return
	}

	MainParams.square_size = 8

	switch MainParams.mode {
	case 1:
		MainParams.mix = .5
		MainParams.N = 25
		MainParams.shuffle_bits = true
	case 2:
		MainParams.mix = .1
		MainParams.N = 15
		MainParams.shuffle_bits = true
	case 3:
		MainParams.mix = .05
		MainParams.N = 10
		MainParams.shuffle_bits = false
	default:
		code = 4
		return
	}

	if MainParams.encode_decode == 1 {

		err = copyXN.CopyXNFile(MainParams.binary_input, MainParams.xN_input, MainParams.N)
		if err != nil {
			return
		}

		err = stir.ShuffleFile(MainParams.xN_input, MainParams.xN_input_stir, MainParams.code_key, MainParams.shuffle_bits)
		if err != nil {
			return
		}

		_ = os.Remove(MainParams.xN_input)

		step1 := exec.Command(
			"ffmpeg", "-y",
			"-i", MainParams.back_filename,
			"-q:a", "0",
			"-map", "a",
			filepath.Join(MainParams.tempDir_audio, "back.mp3"),
			filepath.Join(MainParams.tempDir_bpics, "pic_%8d.bmp"),
			"-threads", strconv.Itoa(runtime.GOMAXPROCS(0)),
		)

		step1.Stderr = os.Stderr
		step1.Stdout = os.Stdout

		err = step1.Run()
		if err != nil {
			return
		}

		if MainParams.max_bpics, err = countFiles(MainParams.tempDir_bpics); err != nil {
			return
		}

		if MainParams.max_bpics <= 0 {
			err = fmt.Errorf("There are no files in directory %s", MainParams.tempDir_bpics)
			return
		}

		MainParams.width_pics, MainParams.height_pics, err = GetPNGDimensions(filepath.Join(MainParams.tempDir_bpics, fmt.Sprintf("pic_%0*d.bmp", 8, 1)))
		if err != nil {
			return
		}

		if (MainParams.width_pics%MainParams.square_size != 0) || (MainParams.height_pics%MainParams.square_size != 0) {
			err = fmt.Errorf("The height and width of the back video should be completely divided by the number %d", MainParams.square_size)
			return
		}

		var fs int64

		fs, err = bin2pics(MainParams)
		if err != nil {
			return
		}

		_ = os.Remove(MainParams.xN_input_stir)

		join_bmps := exec.Command(
			"ffmpeg", "-y",
			"-pattern_type", "glob",
			"-i", filepath.Join(MainParams.tempDir_rpics, "*.bmp"),
			filepath.Join(MainParams.tempDir_res, "video.webm"),
			"-threads", strconv.Itoa(runtime.GOMAXPROCS(0)),
		)

		join_bmps.Stderr = os.Stderr
		join_bmps.Stdout = os.Stdout

		err = join_bmps.Run()
		if err != nil {
			return
		}

		if MainParams.max_rpics, err = countFiles(MainParams.tempDir_rpics); err != nil {
			return
		}

		repetition_rate := int(math.Ceil(float64(MainParams.max_rpics) / float64(MainParams.max_bpics)))

		if repetition_rate > 1 {

			repeat_audio := exec.Command(
				"ffmpeg", "-y",
				"-stream_loop", strconv.Itoa(repetition_rate-1),
				"-i", filepath.Join(MainParams.tempDir_audio, "back.mp3"),
				filepath.Join(MainParams.tempDir_audio, "new_back.mp3"),
				"-threads", strconv.Itoa(runtime.GOMAXPROCS(0)),
			)

			repeat_audio.Stderr = os.Stderr
			repeat_audio.Stdout = os.Stdout

			err = repeat_audio.Run()
			if err != nil {
				return
			}
		} else {
			err = os.Link(filepath.Join(MainParams.tempDir_audio, "back.mp3"), filepath.Join(MainParams.tempDir_audio, "new_back.mp3"))
			if err != nil {
				return
			}
		}

		main_result := exec.Command(
			"ffmpeg", "-y",
			"-i", filepath.Join(MainParams.tempDir_audio, "new_back.mp3"),
			"-i", filepath.Join(MainParams.tempDir_res, "video.webm"),
			// "-map", "0:v",
			// "-map", "1:a",
			// "-c:v", "copy",
			// "-c:a", "copy",
			filepath.Join(MainParams.tempDir_res, "main_video.webm"),
			"-threads", strconv.Itoa(runtime.GOMAXPROCS(0)),
		)

		main_result.Stderr = os.Stderr
		main_result.Stdout = os.Stdout

		err = main_result.Run()
		if err != nil {
			return
		}

		MainParams.video_output = MainParams.binary_input + "_filesize" + strconv.FormatInt(int64(fs), 10) + ".webm"
		_ = os.Remove(MainParams.video_output)
		err = os.Link(filepath.Join(MainParams.tempDir_res, "main_video.webm"), MainParams.video_output)
		if err != nil {
			return
		}
	}

	if MainParams.encode_decode == 2 {

		step1 := exec.Command(
			"ffmpeg", "-y",
			"-i", MainParams.video_input,
			filepath.Join(MainParams.tempDir_rpics, "pic_%8d.bmp"),
			"-threads", strconv.Itoa(runtime.GOMAXPROCS(0)),
		)

		step1.Stderr = os.Stderr
		step1.Stdout = os.Stdout

		err = step1.Run()
		if err != nil {
			return
		}

		err = decodeVideo(MainParams)
		if err != nil {
			return
		}

		err = stir.UnshuffleFile(MainParams.xN_output_stir, MainParams.xN_output, MainParams.code_key, MainParams.shuffle_bits)
		if err != nil {
			return
		}

		_ = os.Remove(MainParams.xN_output_stir)

		MainParams.broken_bits, MainParams.successful_corrections, err = copyXN.RestoreXFile(MainParams.xN_output, MainParams.binary_output, MainParams.N)
		if err != nil {
			return
		}

		_ = os.Remove(MainParams.xN_output)

	}

	code = 0
}
