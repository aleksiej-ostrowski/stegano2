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

package copyXN

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
)

const (
	BUF_SIZE = 4096
)

func CopyXNFile(input_filename, output_filename string, N int) error {
	inputFile, err := os.OpenFile(input_filename, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}

	defer inputFile.Close()

	outputFile, err := os.OpenFile(output_filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	stat, err := inputFile.Stat()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(inputFile)
	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	fileSize_inputFile := stat.Size()

	for iter := 0; iter < N; iter++ {
		_, err := inputFile.Seek(0, 0)
		if err != nil {
			return err
		}

		bytesCopied, err := io.Copy(writer, reader)
		if err != nil {
			return err
		}

		if bytesCopied != fileSize_inputFile {
			return fmt.Errorf("Failed to copy the entire file")
		}

		err = writer.Flush()
		if err != nil {
			return err
		}
	}

	stat, err = outputFile.Stat()
	if err != nil {
		return err
	}

	fileSize_outputFile := stat.Size()
	if fileSize_outputFile != int64(N)*fileSize_inputFile {
		return fmt.Errorf("Failed to copy the entire file %d times", N)
	}

	return nil
}

func acceptablePart(n int) int {
	if n == 2 {
		return n
	}
	return int(math.Round(float64(n) * .67))
}

func checkNbytes(bytes *[][]uint8, n int) ([]uint8, int, int) {

	res := make([]uint8, n)
	cors := 0
	errs := 0

	N_chunks := len(*bytes)
	N_chunks_acceptable := acceptablePart(N_chunks)

	for idx_buf := 0; idx_buf < n; idx_buf++ {
		rs := uint8(0)
		for bit := uint8(0); bit < 8; bit++ {
			al := 0
			for chunk := 0; chunk < N_chunks; chunk++ {
				al += int(((*bytes)[chunk][idx_buf] >> bit) & 1)
			}

			switch {
			case (al > 0) && (al <= (N_chunks - N_chunks_acceptable)):
				cors += 1
			case (al >= N_chunks_acceptable) && (al != N_chunks):
				rs |= 1 << bit
				cors += 1
			case al == N_chunks:
				rs |= 1 << bit
			case al != 0:
				if al > (N_chunks >> 2) {
					rs |= 1 << bit
				}
				errs += 1
			}
		}
		res[idx_buf] = rs
	}

	return res, cors, errs
}

func RestoreXFile(input_filename, output_filename string, N int) (int64, int64, error) {
	errors := int64(0)
	сorrections := int64(0)

	inputFile, err := os.OpenFile(input_filename, os.O_RDONLY, 0600)
	if err != nil {
		return errors, сorrections, err
	}

	defer inputFile.Close()

	outputFile, err := os.OpenFile(output_filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return errors, сorrections, err
	}
	defer outputFile.Close()

	stat, err := inputFile.Stat()
	if err != nil {
		return errors, сorrections, err
	}

	reader := bufio.NewReader(inputFile)
	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	fileSize_inputFile := stat.Size() / int64(N)

	stat, err = outputFile.Stat()
	if err != nil {
		return errors, сorrections, err
	}

	buf := make([][]uint8, N)
	for idx := range buf {
		buf[idx] = make([]uint8, BUF_SIZE)
	}

	bytes_read := int64(0)

	for {

		for chunk := 0; chunk < N-1; chunk++ {
			_, err = inputFile.Seek(fileSize_inputFile*int64(chunk)+bytes_read, 0)
			if err != nil {
				break
			}

			n, err := reader.Read(buf[chunk])
			if n == 0 || err != nil {
				break
			}
		}

		_, err = inputFile.Seek(fileSize_inputFile*(int64(N)-1)+bytes_read, 0)
		if err != nil {
			break
		}

		n, err := reader.Read(buf[N-1])
		if n == 0 || err != nil {
			break
		}

		bytes_read += int64(n)
		save, cors, errs := checkNbytes(&buf, n)
		errors += int64(errs)
		сorrections += int64(cors)

		_, err = writer.Write(save)
		if err != nil {
			break
		}

		if bytes_read == fileSize_inputFile || n < BUF_SIZE {
			break
		}
	}

	if err != nil && err != io.EOF {
		return errors, сorrections, err
	}

	writer.Flush()

	return errors, сorrections, nil
}

/*

func main() {
     _ = CopyXNFile("test.webm", "test.webm_", 4)
     _, _, _ = RestoreXFile("test.webm_", "test.webm2", 4)
}
*/
