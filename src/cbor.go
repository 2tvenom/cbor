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
	v := Vector{1,2,3}

	var buffTest bytes.Buffer
	encoder := cbor.NewEncoder(&buffTest)
	ok, error := encoder.Encode(v)

	if !ok {
		fmt.Errorf("Error decoding %s", error)
	} else {
		fmt.Printf("Variable Hex = % x\n", buffTest.Bytes())
		fmt.Printf("Variable = %v\n", buffTest.Bytes())
	}
}
