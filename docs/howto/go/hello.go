package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Person struct {
	Name string
}

func main() {
	p := &Person{}
	if err := json.NewDecoder(os.Stdin).Decode(p); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("Hello go! My name is", p.Name)
}
