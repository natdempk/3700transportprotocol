package tpl

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestCompression(*testing.T) {
	data, _ := ioutil.ReadFile("mmmm")
	compressed := CompressBytes(data)
	decompressed := DecompressBytes(compressed)
	fmt.Printf("%s\n", decompressed)
	fmt.Printf("%s", data)

	if len(decompressed) == len(data) {
		fmt.Println("victoly")
	}

}
