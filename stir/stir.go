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

package stir

import (
	"bufio"
	"github.com/edsrzf/mmap-go"
	"hash/fnv"
	"io"
	"math/rand"
	"os"
	// "fmt"
)

func hashInt64(code string) int64 {
	hash := fnv.New64a()
	hash.Write([]byte(code))
	return int64(hash.Sum64())
}

func shuffleBits(b uint8, rnd_ *rand.Rand) uint8 {

	bits := make([]uint8, 8)
	for idx := 0; idx < 8; idx++ {
		bits[idx] = (b >> idx) & 1
	}

	indexes := make([]uint8, 8)
	for idx := range indexes {
		indexes[idx] = uint8(idx)
	}

	rnd_.Shuffle(len(indexes), func(i, j int) {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	})

	new_bits := make([]uint8, 8)
	for idx := range indexes {
		new_bits[idx] = bits[indexes[idx]]
	}

	shuffled := uint8(0)
	for idx, el := range new_bits {
		shuffled |= el << idx
	}

	return shuffled
}

func unshuffleBits(b uint8, rnd_ *rand.Rand) uint8 {

	bits := make([]uint8, 8)
	for idx := 0; idx < 8; idx++ {
		bits[idx] = (b >> idx) & 1
	}

	indexes := make([]uint8, 8)
	for idx := range indexes {
		indexes[idx] = uint8(idx)
	}

	rnd_.Shuffle(len(indexes), func(i, j int) {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	})

	new_bits := make([]uint8, 8)
	for idx := range indexes {
		new_bits[indexes[idx]] = bits[idx]
	}

	shuffled := uint8(0)
	for idx, el := range new_bits {
		shuffled |= el << idx
	}

	return shuffled
}

func ShuffleFile(input_file, output_file string, code string, shuffle_bits bool) error {
	inputFile, err := os.OpenFile(input_file, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	fileInfo, err := inputFile.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	mmapData, err := mmap.Map(inputFile, mmap.RDONLY, 0)
	if err != nil {
		return err
	}

	indexes := make([]uint32, fileSize)
	for idx := range indexes {
		indexes[idx] = uint32(idx)
	}

	rnd := rand.New(rand.NewSource(hashInt64(code)))
	rnd.Shuffle(len(indexes), func(i, j int) {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	})

	outputFile, err := os.OpenFile(output_file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	for idx := range indexes {
		el := mmapData[indexes[idx]]

		if shuffle_bits {
			el = shuffleBits(el, rnd)
		}

		err = writer.WriteByte(el)
		if err != nil {
			return err
		}
	}

	writer.Flush()

	if err := mmapData.Unmap(); err != nil {
		return err
	}

	return nil
}

func UnshuffleFile(input_file, output_file string, code string, unshuffle_bits bool) error {
	inputFile, err := os.OpenFile(input_file, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	fileInfo, err := inputFile.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	indexes := make([]uint32, fileSize)
	for idx := range indexes {
		indexes[idx] = uint32(idx)
	}

	rnd := rand.New(rand.NewSource(hashInt64(code)))
	rnd.Shuffle(len(indexes), func(i, j int) {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	})

	outputFile, err := os.OpenFile(output_file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = os.Truncate(output_file, fileSize)
	if err != nil {
		return err
	}

	mmapData_output, err := mmap.Map(outputFile, mmap.RDWR, 0)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(inputFile)

	for idx := range indexes {

		el, err := reader.ReadByte()
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF {
			break
		}

		if unshuffle_bits {
			el = unshuffleBits(el, rnd)
		}

		mmapData_output[indexes[idx]] = el
	}

	mmapData_output.Flush()

	if err := mmapData_output.Unmap(); err != nil {
		return err
	}

	return nil
}

/*

func main() {

    rnd := rand.New(rand.NewSource(101))
    // fmt.Println("=> ", rnd.Intn(10))

    r1 := shuffleBits(117, rnd)
	fmt.Println(r1)

    rnd = rand.New(rand.NewSource(101))
    // fmt.Println("=> ", rnd.Intn(10))

    r2 := unshuffleBits(r1, rnd)
	fmt.Println(r2)

*/

/*

	file_input := "test.webm"
	file_output := file_input + "_"
	file_output2 := file_output + "_"

	err := ShuffleFile(file_input, file_output, "127")
	if err != nil {
		fmt.Println(err)
	}

	err = UnshuffleFile(file_output, file_output2, "127")
	if err != nil {
		fmt.Println(err)
    }

}
*/
