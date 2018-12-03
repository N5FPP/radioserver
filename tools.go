package main

import (
	"bytes"
	"encoding/binary"
)

func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func structToBytes(s interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	binary.Write(buff, binary.LittleEndian, s)

	return buff.Bytes()
}
