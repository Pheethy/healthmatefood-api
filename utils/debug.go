package utils

import (
	"encoding/json"
	"fmt"
	"log"
)

func Debug(data any) {
	bt, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatalf("Debug Failed: %s", err)
	}
	fmt.Println(string(bt))
}

func OutPut(data any) []byte {
	bt, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Debug Failed: %s", err)
	}
	return bt
}