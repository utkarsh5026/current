package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/mybittorrent/bencode"
	"net"
	"os"
	"strconv"
)

func readTorrentFile(fileName string) string {
	contents, err := os.ReadFile(fileName)
	exitIfError(err)
	return string(contents)
}

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
	exitIfError(err)

	decoded = convertNilToEmpty(decoded)
	jsonOutput, err := json.Marshal(decoded)
	exitIfError(err)

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

	conn, handshake, err := bencode.HandShakeWithPeer(*torrentInfo, peerAddress)
	if err != nil {
		return fmt.Errorf("error handshaking with peer: %w", err)
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	peerId := handshake[48:]
	fmt.Printf("Peer ID: %s\n", hex.EncodeToString(peerId))

	return nil
}

func downloadPiece(torrentFile, outputPath string, pieceIdx int) error {
	contents := readTorrentFile(torrentFile)
	parser := bencode.CreateParser(contents)

	torrentInfo, err := parser.ParseTorrent()
	if err != nil {
		return err
	}

	torrentInfo.PrintStats()
	return bencode.DownLoadFile(*torrentInfo, outputPath, pieceIdx)
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
		exitIfError(err)

		parser := bencode.CreateParser(string(contents))
		torrentInfo, err := parser.ParseTorrent()
		exitIfError(err)

		torrentInfo.PrintStats()

	case "peers":
		fileName := os.Args[2]
		err := showPeers(fileName)
		exitIfError(err)

	case "handshake":
		fileName := os.Args[2]
		peerAddress := os.Args[3]

		err := printPeerIdFromHandshake(fileName, peerAddress)
		exitIfError(err)

	case "download_piece":
		outputPPath := os.Args[3]
		fileName := os.Args[4]
		pieceIdx, err := strconv.Atoi(os.Args[5])
		exitIfError(err)

		err = downloadPiece(fileName, outputPPath, pieceIdx)
		exitIfError(err)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}

}

func exitIfError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
