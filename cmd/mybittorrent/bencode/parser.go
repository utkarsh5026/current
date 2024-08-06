package bencode

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

type Parser struct {
	decoder *Decoder
	encoder *Encoder
}

func CreateParser(input string) *Parser {
	return &Parser{
		decoder: NewDecoder(input),
		encoder: &Encoder{},
	}
}

func (p *Parser) Parse() (interface{}, error) {
	return p.decoder.Decode()
}

func (p *Parser) ParseTorrent() (*TorrentInfo, error) {
	value, err := p.Parse()
	if err != nil {
		return nil, err
	}

	dict, ok := value.(map[string]interface{})
	if !ok {
		return nil, InvalidFormat(TypeDict)
	}

	info, ok := dict["info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing info dictionary")
	}

	announce, ok := dict["announce"].(string)
	if !ok {
		return nil, fmt.Errorf("missing announce URL")
	}

	length, ok := info["length"].(int)
	if !ok {
		return nil, fmt.Errorf("missing length")
	}

	pieceLength, ok := info["piece length"].(int)
	if !ok {
		return nil, fmt.Errorf("missing piece length")
	}

	pieces, ok := info["pieces"].(string)
	if !ok {
		return nil, fmt.Errorf("missing pieces")
	}
	hash, err := calculateInfoHash(p, info)
	piecesHashes, err := extractPieceHashes(pieces)

	if err != nil {
		return nil, err
	}

	return &TorrentInfo{
		Announce:    announce,
		Length:      int64(length),
		Info:        info,
		InfoHash:    hash,
		PieceLength: int64(pieceLength),
		PieceHashes: piecesHashes,
	}, nil
}

// calculateInfoHash calculates the SHA-1 hash of the encoded info dictionary.
// The info dictionary is first encoded into a bencoded string, and then the SHA-1 hash is computed.
//
// Parameters:
// - parser: A pointer to the Parser instance containing the encoder.
// - info: A map[string]interface{} representing the info dictionary to be hashed.
//
// Returns:
// - A string containing the hexadecimal representation of the SHA-1 hash.
// - An error if the encoding of the info dictionary fails.
func calculateInfoHash(parser *Parser, info map[string]interface{}) (string, error) {
	encoded, err := parser.encoder.encodeDict(info)
	if err != nil {
		return "", err
	}

	hash := sha1.Sum([]byte(encoded))
	return hex.EncodeToString(hash[:]), nil
}

func extractPieceHashes(pieces string) ([]string, error) {
	if len(pieces)%20 != 0 {
		return nil, fmt.Errorf("invalid pieces length")
	}

	var hashes []string
	for i := 0; i < len(pieces); i += 20 {
		hash := pieces[i : i+20]
		hashes = append(hashes, hex.EncodeToString([]byte(hash)))
	}

	return hashes, nil
}
