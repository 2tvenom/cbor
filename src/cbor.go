package main

import (
	"fmt"
	"cbor"
)

func main() {
	target := map[int]interface{}{1:10,2:"abc",3:30}

	byteBuff, error := cbor.Encode(target)
	if error != nil {
		fmt.Errorf("Error encoding %s", error)
	} else {
		fmt.Printf("Hex map = % x\n", byteBuff)
		fmt.Printf("Dec map = %v\n", byteBuff)
	}
}
