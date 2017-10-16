package cbor

import (
	"bytes"
	"testing"
)

type Range struct {
	Length int
	Align  float32
}

type Vector struct {
	X, Y, Z int
	Range   []Range
	Label   string `tag:"base64"`
}

var v = &Vector{
	X: 10,
	Y: 15,
	Z: 100,
	Range: []Range{
		Range{1, 10},
		Range{223432423, 30},
		Range{3, 41.5},
		Range{174, 55555.2},
	},
	Label: "HoHoHo",
}

var buffTest bytes.Buffer

func TestEncode(t *testing.T) {
	encoder := NewEncoder(&buffTest)
	_, err := encoder.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDecode(t *testing.T) {
	var vd Vector
	encoder := NewEncoder(&buffTest)
	_, err := encoder.Unmarshal(buffTest.Bytes(), &vd)
	if err != nil {
		t.Fatal(err)
	}
}
