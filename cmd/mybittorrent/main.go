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

	case map[string]interface{}:
		if v == nil {
			return map[string]interface{}{}
		}
		for key, elem := range v {
			v[key] = convertNilToEmpty(elem)
		}
	}
	return v
}

func decodeBencodedValue(bencodedValue string) string {
	decoder := bencode.NewDecoder(bencodedValue)
	decoded, err := decoder.Decode()
	if err != nil {
		fmt.Println("Error decoding bencoded value: " + err.Error())
		os.Exit(1)
	}

	decoded = convertNilToEmpty(decoded)
	jsonOutput, err := json.Marshal(decoded)
	if err != nil {
		fmt.Println("Error encoding JSON: " + err.Error())
		os.Exit(1)
	}

	return string(jsonOutput)
}

func main() {
	command := os.Args[1]

	switch command {
	case "decode":
		decodedValue := decodeBencodedValue(os.Args[2])
		fmt.Println(decodedValue)

	case "info":
		fileName := os.Args[2]
		contents, err := os.ReadFile(fileName)

		if err != nil {
			fmt.Println("Error reading file: " + err.Error())
			os.Exit(1)
		}

		parser := bencode.CreateParser(string(contents))
		torrentInfo, err := parser.ParseTorrent()

		if err != nil {
			fmt.Println("Error parsing torrent: " + err.Error())
			os.Exit(1)
		}

		torrentInfo.PrintStats()

	case "peers":
		fileName := os.Args[2]
		contents, err := os.ReadFile(fileName)

		if err != nil {
			fmt.Println("Error reading file: " + err.Error())
			os.Exit(1)
		}

		parser := bencode.CreateParser(string(contents))
		torrentInfo, err := parser.ParseTorrent()

		if err != nil {
			fmt.Println("Error parsing torrent: " + err.Error())
			os.Exit(1)
		}

		trackerResp, err := bencode.CallTracker(*torrentInfo)
		if err != nil {
			fmt.Println("Error calling tracker: " + err.Error())
			os.Exit(1)
		}

		peers, err := bencode.ExtractPeers(trackerResp)
		for _, peer := range peers {
			fmt.Println(peer)
		}

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
