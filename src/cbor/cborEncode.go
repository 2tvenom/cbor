package cbor

import (
	"bytes"

)

type cborEncode struct {
	buff *bytes.Buffer
}

func NewEncoder(buff *bytes.Buffer) (cborEncode){
	return cborEncode{buff}
}

func (encoder *cborEncode) Encode(value interface{}) {

}
