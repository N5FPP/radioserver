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

// region Array Type Converters
func unknownArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	binary.Write(buff, binary.LittleEndian, s)

	return buff.Bytes()
}

func float32ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(float32))
	}

	return buff.Bytes()
}

func float64ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(float64))
	}

	return buff.Bytes()
}

func int16ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(int16))
	}

	return buff.Bytes()
}

func int8ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(int8))
	}

	return buff.Bytes()
}

func complex64ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(complex64))
	}

	return buff.Bytes()
}

func uint8ArrayToBytes(s []interface{}) []uint8 {
	var buff = make([]uint8, len(s))

	for i := 0; i < len(s); i++ {
		buff[i] = s[i].(uint8)
	}

	return buff
}

// endregion

func arrayToBytes(s []interface{}) []uint8 {
	if len(s) > 0 {
		var elem = s[0]
		switch _ := elem.(type) {
		case float32:
			return float32ArrayToBytes(s)
		case float64:
			return float64ArrayToBytes(s)
		case complex64:
			return complex64ArrayToBytes(s)
		case uint8:
			return uint8ArrayToBytes(s)
		case int8:
			return int8ArrayToBytes(s)
		case int16:
			return int16ArrayToBytes(s)
		default:
			return unknownArrayToBytes(s)
		}
	}
	return make([]uint8, 0)
}
