package main

import (
	"fmt"

	"github.com/chenguan1/msmap/model"
)

func main() {
	ds := &model.Dataset{}
	ms := &model.Mapset{}
	fmt.Printf("%#v\n", ds)
	fmt.Printf("%#v\n", ms)
}
