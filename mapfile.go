package main

import (
	"text/template"
	"os"
)

type Persion struct {
	Name string
	Age  int
}

func test(){
	muban := "hello {{.Name}}, your are {{.Age}} years old."
	tmpl, err := template.New("test").Parse(muban)
	if err != nil{
		panic(err)
	}

	gray := Persion{
		Name: "gray",
		Age : 30,
	}

	tmpl.Execute(os.Stdout, gray)
}