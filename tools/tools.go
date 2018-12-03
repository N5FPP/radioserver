package tools

import (
	"bytes"
	"encoding/binary"
)

func Min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func StructToBytes(s interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	binary.Write(buff, binary.LittleEndian, s)

	return buff.Bytes()
}

// region Array Type Converters
func UnknownArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	binary.Write(buff, binary.LittleEndian, s)

	return buff.Bytes()
}

func Float32ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(float32))
	}

	return buff.Bytes()
}

func Float64ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(float64))
	}

	return buff.Bytes()
}

func Int16ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(int16))
	}

	return buff.Bytes()
}

func Int8ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(int8))
	}

	return buff.Bytes()
}

func Complex64ArrayToBytes(s []interface{}) []uint8 {
	var buff = new(bytes.Buffer)

	for i := 0; i < len(s); i++ {
		binary.Write(buff, binary.LittleEndian, s[i].(complex64))
	}

	return buff.Bytes()
}

func UInt8ArrayToBytes(s []interface{}) []uint8 {
	var buff = make([]uint8, len(s))

	for i := 0; i < len(s); i++ {
		buff[i] = s[i].(uint8)
	}

	return buff
}

// endregion

func ArrayToBytes(s []interface{}) []uint8 {
	var va interface{}
	if len(s) > 0 {
		var elem = s[0]
		switch v := elem.(type) {
		case float32:
			return Float32ArrayToBytes(s)
		case float64:
			return Float64ArrayToBytes(s)
		case complex64:
			return Complex64ArrayToBytes(s)
		case uint8:
			return UInt8ArrayToBytes(s)
		case int8:
			return Int8ArrayToBytes(s)
		case int16:
			return Int16ArrayToBytes(s)
		default:
			va = v
			return UnknownArrayToBytes(s)
		}
	}
	va = va
	return make([]uint8, 0)
}
