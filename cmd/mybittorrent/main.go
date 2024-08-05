package main

import (
	"encoding/json"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"

	// Uncomment this line to pass the first stage
	// "encoding/json"
	"fmt"
	"os"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func convertNilToEmpty(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		if v == nil {
			return []interface{}{}
		}
		for i, elem := range v {
			v[i] = convertNilToEmpty(elem)
		}
	}
	return v
}

func main() {
	command := os.Args[1]

	if command == "decode" {

		bencodedValue := os.Args[2]

		decoder := bencode.NewDecoder(bencodedValue)
		decoded, err := decoder.Decode()
		decoded = convertNilToEmpty(decoded)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Printf("%s\n", jsonOutput)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
