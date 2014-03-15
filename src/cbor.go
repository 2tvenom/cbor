package main

import (
	"fmt"
	"cbor"
	"bytes"
)

type Vector struct {
	X, Y, Z int
}

func main() {
	var v float32

	v = 12.3

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

	var vd float32

	ok, err := encoder.Unmarshal(buffTest.Bytes(), &vd)

	if !ok {
		fmt.Printf("Error Unmarshal %s", err)
		return
	}

	fmt.Printf("%v", vd)
}
