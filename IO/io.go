package main

import (
	"io/ioutil"
	"log"
)

func main() {
	message := []byte("Hello, Gophers!")
	err := ioutil.WriteFile("testdata/hello", message, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
