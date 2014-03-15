package main

import (
	"fmt"
	"cbor"
	"bytes"
)

type Vector struct {
	X, Y, Z int
	Range []Range
	Label string
}

type Range struct {
	Length int
	Align float32
}

func main() {
	v := &Vector {
		X: 10,
		Y: 15,
		Z: 100,
		Range: []Range{
			Range {1,10},
			Range {223432423,30},
			Range {3,41.5},
			Range {174,55555.2},
		},
		Label: "HoHoHo",
	}

	var buffTest bytes.Buffer
	encoder := cbor.NewEncoder(&buffTest)
	ok, error := encoder.Marshal(v)

	if !ok {
		fmt.Errorf("Error decoding %s", error)
	} else {
		fmt.Printf("Variable Hex = % x\n", buffTest.Bytes())
		fmt.Printf("Variable = %v\n", buffTest.Bytes())
	}
	fmt.Printf("-----------------\n")

	var vd Vector

	ok, err := encoder.Unmarshal(buffTest.Bytes(), &vd)

	if !ok {
		fmt.Printf("Error Unmarshal %s", err)
		return
	}

	fmt.Printf("%v", vd)
}
