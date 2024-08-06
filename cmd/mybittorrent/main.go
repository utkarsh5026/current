package main

import (
	"encoding/hex"
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

func showPeers(fileName string) error {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	parser := bencode.CreateParser(string(contents))
	torrentInfo, err := parser.ParseTorrent()
	if err != nil {
		return fmt.Errorf("error parsing torrent: %w", err)
	}

	trackerResp, err := bencode.CallTracker(*torrentInfo)
	if err != nil {
		return fmt.Errorf("error calling tracker: %w", err)
	}

	peers, err := bencode.ExtractPeers(trackerResp)
	if err != nil {
		return fmt.Errorf("error extracting peers: %w", err)
	}

	for _, peer := range peers {
		fmt.Println(peer)
	}

	return nil
}

func printPeerIdFromHandshake(fileName string, peerAddress string) error {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	parser := bencode.CreateParser(string(contents))
	torrentInfo, err := parser.ParseTorrent()
	if err != nil {
		return fmt.Errorf("error parsing torrent: %w", err)
	}

	handshake, err := bencode.HandShakeWithPeer(*torrentInfo, peerAddress)
	if err != nil {
		return fmt.Errorf("error handshaking with peer: %w", err)
	}

	peerId := handshake[48:]
	fmt.Printf("Peer ID: %s\n", hex.EncodeToString(peerId))

	return nil

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
		err := showPeers(fileName)

		if err != nil {
			fmt.Println("Error showing peers: " + err.Error())
			os.Exit(1)
		}

	case "handshake":
		fileName := os.Args[2]
		peerAddress := os.Args[3]

		err := printPeerIdFromHandshake(fileName, peerAddress)

		if err != nil {
			fmt.Println("Error printing peer ID: " + err.Error())
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}
