package cbor

import (
	"fmt"
	"encoding/binary"
	"bytes"
	"unicode/utf8"
//	"reflect"
	"reflect"
)

const (
	majorOffset = 5
	additionalMax = 23
	additionalTypeIntFalse byte = 20
	additionalTypeIntTrue byte = 21
	additionalTypeIntNull byte = 22
	additionalTypeIntUndefined byte = 23
	additionalTypeIntUint8 byte = 24
	additionalTypeIntUint16 byte = 25
	additionalTypeIntUint32 byte = 26
	additionalTypeIntUint64 byte = 27
	additionalTypeFloat16 byte = 25
	additionalTypeFloat32 byte = 26
	additionalTypeFloat64 byte = 27
	additionalTypeBreak byte = 31
	float64PackType = "d"
)

const (
	majorTypeUnsignedInt byte = iota << majorOffset
	majorTypeInt
	majorTypeByteString
	majorTypeUtf8String
	majorTypeArray
	majorTypeMap
	majorTypeTags
	majorTypeSimpleAndFloat
)

func Encode(variable interface{}) ([]byte, error) {
	switch reflect.TypeOf(variable).Kind() {
	case reflect.Int:
		return encodeNumber(variable.(int))
	case reflect.String:
		return encodeString(variable.(string))
	case reflect.Array, reflect.Slice, reflect.Map:
		return encodeArray(variable)
	}

	return nil, nil
}

/**
	Encode array/map to CBOR binary string
 */
func encodeArray(variable interface{}) ([]byte, error) {
	majorType := majorTypeArray
	inputSlice := reflect.ValueOf(variable)
	length := inputSlice.Len()

	if inputSlice.Kind() == reflect.Map {
		majorType = majorTypeMap
	}

	buff, err := packNumber(majorType, uint64(length))

	//array slice encode
	if inputSlice.Kind() != reflect.Map {
		for i:=0; i < inputSlice.Len(); i++ {
			elementBuff, err := Encode(inputSlice.Index(i).Interface())

			if err != nil {
				return nil, err
			}

			buff = append(buff, elementBuff...)
		}
	} else {
		//map encode
		for _, key := range inputSlice.MapKeys() {
			keyBuff, keyErr := Encode(key.Interface())

			if keyErr != nil {
				return nil, keyErr
			}

			buff = append(buff, keyBuff...)

			elementBuff, err := Encode(inputSlice.MapIndex(key).Interface())

			if err != nil {
				return nil, err
			}

			buff = append(buff, elementBuff...)
		}
	}

	if err != nil {
		return nil, err
	}

	return buff, nil
}

/**
	Encode string to CBOR binary string
 */
func encodeString(variable string) ([]byte, error) {
	byteBuf := []byte(variable)

	majorType := majorTypeUtf8String

	if !utf8.Valid(byteBuf) {
		majorType = majorTypeByteString
	}

	initByte, err := packNumber(majorType, uint64(len(byteBuf)))

	if err != nil {
		return []byte{}, err
	}

	return append(initByte, byteBuf...), nil
}

/**
	Encode integer to CBOR binary string
 */
func encodeNumber(variable int) ([]byte, error) {
	var majorType = majorTypeUnsignedInt

	var unsignedVariable uint64

	if variable < 0 {
		majorType = majorTypeInt
		unsignedVariable = uint64(-(variable + 1))
	} else {
		unsignedVariable = uint64(variable)
	}

	byteArr, err := packNumber(majorType, unsignedVariable)
	return byteArr, err
}

/**
	Pack number helper
 */
func packNumber(majorType byte, number uint64) ([]byte, error){
	if number < additionalMax {
		return packInitByte(majorType, byte(number))
	}

	additionInfo := intTypeToCborType(number)

	initByte, err := packInitByte(majorType, additionInfo)

	if err != nil {
		return []byte{}, err
	}

	var packedInfo []byte
	var errPack error

	switch
	{
	case additionInfo == additionalTypeIntUint8:
		packedInfo, errPack = pack(uint8(number))
	case additionInfo == additionalTypeIntUint16:
		packedInfo, errPack = pack(uint16(number))
	case additionInfo == additionalTypeIntUint32:
		packedInfo, errPack = pack(uint32(number))
	case additionInfo == additionalTypeIntUint64:
		packedInfo, errPack = pack(uint64(number))
	}

	if errPack != nil {
		return nil, errPack
	}

	return append(initByte, packedInfo...), nil
}

/**
	Helper for packing Go objects. Like in C, PHP function pack()
 */
func pack(packVariable interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, packVariable)

	if err != nil {
		return nil, fmt.Errorf("Cant pack init byte. %s", err)
	}

	return buf.Bytes(), nil
}

/**
	Pack initial bye
*/
func packInitByte(majorType byte, additionalInfo byte) ([]byte, error) {
	return pack(majorType | additionalInfo)
}

/**
	Get CBOR additional info type for number
*/
func intTypeToCborType(number uint64) (byte) {
	switch {
	case number < 256:
		return additionalTypeIntUint8
	case number < 65536:
		return additionalTypeIntUint16
	case number < 4294967296:
		return additionalTypeIntUint32
	default:
		return additionalTypeIntUint64
	}
}
