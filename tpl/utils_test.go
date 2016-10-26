package tpl

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestCompression(*testing.T) {
	data, _ := ioutil.ReadFile("mmmm")
	compressed := CompressBytes(data)
	fmt.Println(len(compressed))
	decompressed := DecompressBytes(compressed, false)
	_ = decompressed
	fmt.Println(len(data), len(decompressed))
	// fmt.Printf("%s\n", decompressed)
	// fmt.Printf("%s\n", data)

	if len(decompressed) == len(data) {
		fmt.Println("victoly")
	} else {
		fmt.Println("failure")
	}

}
