package bencode

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

type TorrentInfo struct {
	Announce    string
	Length      int64
	Info        map[string]interface{}
	InfoHash    string
	PieceLength int64
	PieceHashes []string
}

func (t TorrentInfo) PrintStats() {
	fmt.Printf("Tracker URL: %v\n", t.Announce)
	fmt.Printf("Length: %v\n", t.Length)
	fmt.Printf("Info Hash: %v\n", t.InfoHash)
	fmt.Printf("Piece Length: %v\n", t.PieceLength)
	fmt.Println("Piece Hashes:")

	for _, hash := range t.PieceHashes {
		fmt.Printf("\t%v\n", hash)
	}
}

func verifyPiece(piece []byte, expectedHash []byte) bool {
	hash := sha1.Sum(piece)
	return hex.EncodeToString(hash[:]) == string(expectedHash)
}
