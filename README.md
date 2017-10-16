# CBOR encoder for Go

Decoder/encoder from Go data to CBOR binary string. This code has been developed and maintained by Ven at March 2014.

CBOR is an object representation format defined by the [IETF](http://ietf.org).
The [specification](http://tools.ietf.org/html/rfc7049)
has recently been approved as an IETF Standards-Track specification
and has been published as RFC 7049.

## Usage
```go
package main

import (
	"fmt"
	"cbor"
	"bytes"
)
//custom struct
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

	//create encoder and marshal
	var buffTest bytes.Buffer
	encoder := cbor.NewEncoder(&buffTest)
	ok, error := encoder.Marshal(v)
	//check binary string
	if !ok {
		fmt.Errorf("Error decoding %s", error)
	} else {
		fmt.Printf("Variable Hex = % x\n", buffTest.Bytes())
		fmt.Printf("Variable = %v\n", buffTest.Bytes())
	}
	fmt.Printf("-----------------\n")

	//unmarshal binary string to new struct
	var vd Vector
	ok, err := encoder.Unmarshal(buffTest.Bytes(), &vd)

	if !ok {
		fmt.Printf("Error Unmarshal %s", err)
		return
	}
	//output
	fmt.Printf("%v", vd)
}
```

## Compatibility

Checked with [PHP extension](https://github.com/2tvenom/CBOREncode) in encode and decode

## Known issues

- Not support tags. 6 major type *(in future)*
- Not support 16  floats encoding
- Not decode nil (null) vars
- Encode does't support indefinite-length values.
