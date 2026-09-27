package bitio

import "testing"

const (
	VALUE_LOW     = 0xA5
	VALUE_HIGH    = 0x3C
	CNT_BITS_TEST = 13
	CNT_BYTES_REF = 2
)

func TestGetPut(test *testing.T) {
	source := []byte{VALUE_LOW, VALUE_HIGH}
	target := make([]byte, len(source))
	cntBits := CntBits(len(source))
	for indBit := 0; indBit < cntBits; indBit++ {
		Put(target, indBit, Get(source, indBit))
	}
	if target[0] != VALUE_LOW ||
		target[1] != VALUE_HIGH {
		test.Fatalf("поток %v", target)
	}
}

func TestCounts(test *testing.T) {
	isRight := CntBytes(CNT_BITS_TEST) ==
		CNT_BYTES_REF &&
		CntBytes(0) == 0 &&
		CntBits(CNT_BYTES_REF) ==
			CNT_BYTES_REF*CNT_BITS_BYTE
	if !isRight {
		test.Fatalf("счёт битов неверен")
	}
}
