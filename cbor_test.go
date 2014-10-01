package cbor_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sharpner/cbor"
)

var _ = Describe("Cbor", func() {

	type Range struct {
		Length int
		Align  float32
	}

	//custom struct
	type Vector struct {
		X, Y, Z int
		Range   []Range
		Label   string
	}

	Context("Test imported from original code", func() {
		v := &Vector{
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

		It("Should encode", func() {
			encoder := cbor.NewEncoder(&buffTest)
			_, err := encoder.Marshal(v)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should decode", func() {
			var vd Vector
			encoder := cbor.NewEncoder(&buffTest)
			_, err := encoder.Unmarshal(buffTest.Bytes(), &vd)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
