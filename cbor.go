package cbor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

const (
	majorOffset                     = 5
	additionalWipe                  = 7 << majorOffset
	majorWipe                       = 31
	additionalMax                   = 23
	additionalTypeIntFalse     byte = 20
	additionalTypeIntTrue      byte = 21
	additionalTypeIntNull      byte = 22
	additionalTypeIntUndefined byte = 23
	additionalTypeIntUint8     byte = 24
	additionalTypeIntUint16    byte = 25
	additionalTypeIntUint32    byte = 26
	additionalTypeIntUint64    byte = 27
	additionalTypeFloat16      byte = 25
	additionalTypeFloat32      byte = 26
	additionalTypeFloat64      byte = 27
	additionalTypeBreak        byte = 31
)

const (
	majorTypeUnsignedInt    byte = iota << majorOffset // Major type 0
	majorTypeInt                                       // Major type 1
	majorTypeByteString                                // Major type 2
	majorTypeUtf8String                                // Major type 3
	majorTypeArray                                     // Major type 4
	majorTypeMap                                       // Major type 5
	majorTypeTags                                      // Major type 6
	majorTypeSimpleAndFloat                            // Major type 7
)

var additionalLength = map[byte]byte{
	additionalTypeIntUint8:  1,
	additionalTypeIntUint16: 2,
	additionalTypeIntUint32: 4,
	additionalTypeIntUint64: 8,
}

var reflectToCbor = map[reflect.Kind][]byte{
	reflect.Bool:    []byte{majorTypeSimpleAndFloat},
	reflect.Int:     []byte{majorTypeInt, majorTypeUnsignedInt},
	reflect.Float32: []byte{majorTypeSimpleAndFloat},
	reflect.Float64: []byte{majorTypeSimpleAndFloat},
	reflect.Array:   []byte{majorTypeArray},
	reflect.Map:     []byte{majorTypeMap},
	reflect.Slice:   []byte{majorTypeArray},
	reflect.String:  []byte{majorTypeByteString, majorTypeUtf8String},
	reflect.Struct:  []byte{majorTypeMap},
}

var typeLabel = map[byte]string{
	majorTypeUnsignedInt:    "Unsignet int",
	majorTypeInt:            "Int",
	majorTypeByteString:     "Byte string",
	majorTypeUtf8String:     "UTF-8 string",
	majorTypeArray:          "Array",
	majorTypeMap:            "Map",
	majorTypeSimpleAndFloat: "Float",
}

type cborEncode struct {
	buff *bytes.Buffer
}

// NewEncoder creates new encoder object
func NewEncoder(buff *bytes.Buffer) cborEncode {
	return cborEncode{buff}
}

func (encoder *cborEncode) Marshal(value interface{}) (bool, error) {
	encoder.buff.Reset()

	ok, err := encoder.encodeValue(value)
	if !ok {
		return false, err
	}

	return true, nil
}

func (encoder *cborEncode) encodeValue(variable interface{}) (bool, error) {
	if variable == nil {
		return encoder.encodeNil()
	}

	switch reflect.TypeOf(variable).Kind() {
	case reflect.Ptr:
		return encoder.encodeValue(reflect.ValueOf(variable).Elem().Interface())
	case reflect.Int:
		return encoder.encodeNumber(variable.(int))
	case reflect.String:
		return encoder.encodeString(variable.(string))
	case reflect.Array, reflect.Slice:
		return encoder.encodeArray(variable)
	case reflect.Map:
		return encoder.encodeMap(variable)
	case reflect.Struct:
		return encoder.encodeStruct(variable)
	case reflect.Bool:
		return encoder.encodeBool(variable.(bool))
	case reflect.Float32:
		return encoder.encodeFloat(variable.(float32), additionalTypeFloat32)
	case reflect.Float64:
		return encoder.encodeFloat(variable.(float64), additionalTypeFloat64)
	}

	return true, nil
}

func (encoder *cborEncode) Unmarshal(data []byte, v interface{}) (bool, error) {
	if len(data) == 0 {
		return false, fmt.Errorf("Empty input byte array")
	}

	reflectedValue := reflect.ValueOf(v)

	if reflectedValue.Kind() != reflect.Ptr {
		return false, fmt.Errorf("Input value must be ptr")
	}

	encoder.buff.Reset()
	_, err := encoder.buff.Write(data)

	if err != nil {
		return false, err
	}

	return encoder.decode(reflectedValue.Elem())
}

func checkReflectWithCbor(reflecType reflect.Type, cborType byte) bool {
	if value, ok := reflectToCbor[reflecType.Kind()]; ok {
		for _, v := range value {
			if cborType == v {
				return true
			}
		}
	}
	return false
}

func (encoder *cborEncode) decode(v reflect.Value) (bool, error) {
	firstElem, err := encoder.buff.ReadByte()

	if err != nil {
		return false, err
	}

	majorType := firstElem & additionalWipe

	ok := checkReflectWithCbor(v.Type(), majorType)

	if !ok {
		return false, fmt.Errorf("Can't convert %s to %s", v, typeLabel[majorType])
	}

	headerAdditionInfo := firstElem & majorWipe

	var dataLength int
	if length, ok := additionalLength[headerAdditionInfo]; ok {
		dataLength += int(length)
	}

	buff := make([]byte, dataLength)
	_, err = encoder.buff.Read(buff)

	if err != nil {
		return false, err
	}

	switch majorType {
	case majorTypeUnsignedInt, majorTypeInt:
		number, err := decodeInt(headerAdditionInfo, buff)

		if err != nil {
			return false, err
		}

		if majorType == majorTypeInt {
			number = -(number + 1)
		}

		v.Set(reflect.ValueOf(number))

		return true, nil
	case majorTypeByteString, majorTypeUtf8String:
		length, err := decodeInt(headerAdditionInfo, buff)

		if err != nil {
			return false, err
		}

		stringBuff := make([]byte, length)
		_, err = encoder.buff.Read(stringBuff)

		if err != nil {
			return false, err
		}

		v.Set(reflect.ValueOf(string(stringBuff)))

		return true, nil
	case majorTypeArray:
		array_cap, err := decodeInt(headerAdditionInfo, buff)

		if err != nil {
			return false, err
		}

		v.Set(reflect.MakeSlice(v.Type(), array_cap, array_cap))

		for i := 0; i < array_cap; i++ {
			ok, err := encoder.decode(v.Index(i))
			if !ok {
				return false, err
			}
		}

		return true, nil
	case majorTypeMap:
		map_length, err := decodeInt(headerAdditionInfo, buff)

		if err != nil {
			return false, err
		}

		if v.Kind() == reflect.Map {
			v.Set(reflect.MakeMap(v.Type()))

			for i := 0; i < map_length; i++ {
				key := reflect.New(v.Type().Key())

				ok, err := encoder.decode(key.Elem())
				if !ok {
					return false, err
				}

				value := reflect.New(v.Type().Elem())

				ok, err = encoder.decode(value.Elem())
				if !ok {
					return false, err
				}

				v.SetMapIndex(key.Elem(), value.Elem())
			}
		} else if v.Kind() == reflect.Struct {
			v.Set(reflect.New(v.Type()).Elem())

			structFieldList := []string{}

			for i := 0; i < v.NumField(); i++ {
				structFieldList = append(structFieldList, v.Type().Field(i).Name)
			}

			for i := 0; i < map_length; i++ {
				var key string

				ok, err := encoder.decode(reflect.ValueOf(&key).Elem())
				if !ok {
					return false, err
				}

				ok, fieldName := lookupField(key, structFieldList)

				if !ok {
					return false, fmt.Errorf("Field %s not strut %v", key, v)
				}

				ok, err = encoder.decode(v.FieldByName(fieldName))

				if !ok {
					return false, err
				}
			}
		}
		return true, nil
	case majorTypeTags:
		return false, fmt.Errorf("Tags not support")
	case majorTypeSimpleAndFloat:
		switch headerAdditionInfo {
		case additionalTypeIntFalse:
			if v.Kind() != reflect.Bool {
				return false, fmt.Errorf("Can convert %s to bool", v.Type())
			}
			v.Set(reflect.ValueOf(false))
			return true, nil
		case additionalTypeIntTrue:
			if v.Kind() != reflect.Bool {
				return false, fmt.Errorf("Can convert %s to bool", v.Type())
			}
			v.Set(reflect.ValueOf(true))
			return true, nil
		case additionalTypeIntNull:
			return true, nil
		case additionalTypeFloat16:
			return true, fmt.Errorf("Float16 decode not support")
		case additionalTypeFloat32:
			if v.Kind() != reflect.Float32 {
				return false, fmt.Errorf("Can convert %s to float32", v.Type())
			}
			var out float32
			err := unpack(buff, &out)

			if err != nil {
				return false, err
			}

			if v.Kind() == reflect.Float32 {
				v.Set(reflect.ValueOf(&out).Elem())
			} else if v.Kind() == reflect.Float64 {
				out64 := float64(out)
				v.Set(reflect.ValueOf(&out64).Elem())
			} else {
				return false, fmt.Errorf("Can convert %s to float32", v.Type())
			}

			return true, nil
		case additionalTypeFloat64:
			var out float64
			err := unpack(buff, &out)

			if err != nil {
				return false, err
			}

			if v.Kind() == reflect.Float64 {
				v.Set(reflect.ValueOf(&out).Elem())
			} else if v.Kind() == reflect.Float32 {
				out32 := float32(out)
				v.Set(reflect.ValueOf(&out32).Elem())
			} else {
				return false, fmt.Errorf("Can convert %s to float64", v.Type())
			}

			return true, nil
		}
	}

	return true, nil
}

func lookupField(field string, fieldList []string) (bool, string) {
	field = strings.ToLower(field)

	for _, currentField := range fieldList {
		if field != strings.ToLower(currentField) {
			continue
		}
		return true, currentField
	}
	return false, ""
}

//decode int
func decodeInt(headerAdditionInfo byte, buff []byte) (int, error) {
	if headerAdditionInfo <= additionalMax {
		return int(headerAdditionInfo), nil
	}

	var number int
	var err error

	switch headerAdditionInfo {
	case additionalTypeIntUint8:
		return int(buff[0]), nil
	case additionalTypeIntUint16:
		var out uint16
		err = unpack(buff, &out)
		number = int(out)
	case additionalTypeIntUint32:
		var out uint32
		err = unpack(buff, &out)
		number = int(out)
	default:
		var out uint64
		err = unpack(buff, &out)
		number = int(out)
	}

	if err != nil {
		return 0, err
	}

	return number, nil
}

func unpack(byteBuff []byte, target interface{}) error {
	buf := bytes.NewReader(byteBuff)
	err := binary.Read(buf, binary.BigEndian, target)
	return err
}

/**
Encoding float32/float64
*/
func (encoder *cborEncode) encodeFloat(number interface{}, additionalFloatType byte) (bool, error) {
	majorType := majorTypeSimpleAndFloat

	initByte, err := packInitByte(majorType, additionalFloatType)

	if err != nil {
		return false, err
	}

	_, err = encoder.buff.Write(initByte)

	if err != nil {
		return false, err
	}

	var packedInfo []byte
	var errPack error

	switch additionalFloatType {
	case additionalTypeFloat32:
		packedInfo, errPack = pack(number.(float32))
	case additionalTypeFloat64:
		packedInfo, errPack = pack(number.(float64))
	default:
		packedInfo, errPack = nil, nil
	}

	if errPack != nil {
		return false, errPack
	}

	_, err = encoder.buff.Write(packedInfo)

	if err != nil {
		return false, err
	}

	return true, nil
}

/**
encoding nil
*/
func (encoder *cborEncode) encodeNil() (bool, error) {
	byteArr, err := packInitByte(majorTypeSimpleAndFloat, additionalTypeIntNull)
	if err != nil {
		return false, err
	}
	encoder.buff.Write(byteArr)
	return true, nil
}

/**
encode
*/
func (encoder *cborEncode) encodeBool(variable bool) (bool, error) {
	varType := additionalTypeIntFalse
	if variable {
		varType = additionalTypeIntTrue
	}

	buff, err := packInitByte(majorTypeSimpleAndFloat, varType)

	if err != nil {
		return false, err
	}

	_, err = encoder.buff.Write(buff)

	if err != nil {
		return false, err
	}

	return true, nil
}

// Encode array to CBOR binary string
func (encoder *cborEncode) encodeArray(variable interface{}) (bool, error) {
	majorType := majorTypeArray
	inputSlice := reflect.ValueOf(variable)
	length := inputSlice.Len()

	buff, err := packNumber(majorType, uint64(length))

	if err != nil {
		return false, err
	}

	_, err = encoder.buff.Write(buff)

	if err != nil {
		return false, err
	}
	//array slice encode
	for i := 0; i < inputSlice.Len(); i++ {
		ok, err := encoder.encodeValue(inputSlice.Index(i).Interface())
		if !ok {
			return false, err
		}
	}

	return true, nil
}

//ecnode struct
func (encoder *cborEncode) encodeStruct(variable interface{}) (bool, error) {
	majorType := majorTypeMap

	inputStructValue := reflect.ValueOf(variable)
	inputStructType := inputStructValue.Type()

	length := inputStructValue.NumField()

	publicRange := 0

	for i := 0; i < length; i++ {
		fieldType := inputStructType.Field(i)
		if fieldType.PkgPath != "" {
			continue
		}
		publicRange++
	}

	if publicRange == 0 {
		return false, fmt.Errorf("Struct%v not have public fields", variable)
	}

	buff, err := packNumber(majorType, uint64(publicRange))

	if err != nil {
		return false, err
	}

	_, err = encoder.buff.Write(buff)

	if err != nil {
		return false, err
	}

	//struct encode
	for i := 0; i < length; i++ {
		fieldType := inputStructType.Field(i)
		if fieldType.PkgPath != "" {
			continue
		}

		switch fieldType.Tag.Get("tag") {
		case "":
		case "base64":
			buff, err := packNumber(majorTypeTags, uint64(34))
			if err != nil {
				return false, err
			}
			_, err = encoder.buff.Write(buff)

			if err != nil {
				return false, err
			}
		}

		keyOk, keyErr := encoder.encodeValue(strings.ToLower(fieldType.Name))

		if !keyOk {
			return false, keyErr
		}

		elementOk, elemErr := encoder.encodeValue(inputStructValue.Field(i).Interface())

		if !elementOk {
			return false, elemErr
		}
	}

	return true, nil
}

// Encode map to CBOR binary string
func (encoder *cborEncode) encodeMap(variable interface{}) (bool, error) {
	majorType := majorTypeMap
	inputSlice := reflect.ValueOf(variable)
	length := inputSlice.Len()

	buff, err := packNumber(majorType, uint64(length))

	if err != nil {
		return false, err
	}

	_, err = encoder.buff.Write(buff)
	if err != nil {
		return false, err
	}

	//map encode
	for _, key := range inputSlice.MapKeys() {
		ok, keyErr := encoder.encodeValue(key.Interface())

		if !ok {
			return false, keyErr
		}

		ok, elemErr := encoder.encodeValue(inputSlice.MapIndex(key).Interface())

		if !ok {
			return false, elemErr
		}
	}

	return true, nil
}

/**
Encode string to CBOR binary string
*/
func (encoder *cborEncode) encodeString(variable string) (bool, error) {
	byteBuf := []byte(variable)

	majorType := majorTypeUtf8String

	if !utf8.Valid(byteBuf) {
		majorType = majorTypeByteString
	}

	initByte, err := packNumber(majorType, uint64(len(byteBuf)))

	if err != nil {
		return false, err
	}

	_, err = encoder.buff.Write(initByte)

	if err != nil {
		return false, err
	}

	_, err = encoder.buff.Write(byteBuf)

	if err != nil {
		return false, err
	}

	return true, nil
}

/**
Encode integer to CBOR binary string
*/
func (encoder *cborEncode) encodeNumber(variable int) (bool, error) {
	var majorType = majorTypeUnsignedInt

	var unsignedVariable uint64

	if variable < 0 {
		majorType = majorTypeInt
		unsignedVariable = uint64(-(variable + 1))
	} else {
		unsignedVariable = uint64(variable)
	}

	byteArr, err := packNumber(majorType, unsignedVariable)
	if err != nil {
		return false, err
	}

	_, err = encoder.buff.Write(byteArr)

	if err != nil {
		return false, err
	}

	return true, err
}

/**
Pack number helper
*/
func packNumber(majorType byte, number uint64) ([]byte, error) {
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

	switch additionInfo {
	case additionalTypeIntUint8:
		packedInfo, errPack = pack(uint8(number))
	case additionalTypeIntUint16:
		packedInfo, errPack = pack(uint16(number))
	case additionalTypeIntUint32:
		packedInfo, errPack = pack(uint32(number))
	default:
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
func intTypeToCborType(number uint64) byte {
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
