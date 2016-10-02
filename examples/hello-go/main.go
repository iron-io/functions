package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	dec := json.NewDecoder(os.Stdin)

	var v map[string]interface{}
	if err := dec.Decode(&v); err != nil {
		log.Println("Hello world!")
		return
	}
	if v, ok := v["name"]; ok {
		log.Println(fmt.Sprintf("Hello %s!", v))
	} else {
		log.Println("Hello world!")
	}

}
