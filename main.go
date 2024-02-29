package main

import (
	"fmt"
	"os"
)

func main() {
	_ = os.Mkdir("./public", 0755)
	f, err := os.Create("./public/index.html")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString("<h1>Hello, World!</h1>")

	fmt.Println("generate success!")
}
