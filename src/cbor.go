package main

import (
	"fmt"
	"cbor"
	"io/ioutil"
)

func main() {
	buff, _ := ioutil.ReadFile("/tmp/cbor")

	decodedData, error := cbor.Decode(&buff)
	if error != nil {
		fmt.Errorf("Error decoding %s", error)
	} else {
		fmt.Printf("Variable = %v\n", decodedData)
	}
}
