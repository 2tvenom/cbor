package main

import (
	"fmt"
	"cbor"
//	"io/ioutil"
//	"bytes"
)

type Vector struct {
	X, Y, Z int
}

func main() {
//	var buffTest bytes.Buffer
//
//	encoder := cbor.NewEncoder(&buffTest)
//	encoder.Encode(Vector{1,2,3})
	v := Vector{1,2,3}
	buff, error := cbor.Encode(v)

	if error != nil {
		fmt.Errorf("Error decoding %s", error)
	} else {
		fmt.Printf("Variable Hex = % x\n", buff)
		fmt.Printf("Variable = %v\n", buff)
	}
}
