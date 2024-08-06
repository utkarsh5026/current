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

	hash, err := CalculateInfoHash(p, info)

	if err != nil {
		return nil, err
	}

	return &TorrentInfo{
		Announce: announce,
		Length:   int64(length),
		Info:     info,
		InfoHash: hash,
	}, nil
}

// CalculateInfoHash calculates the SHA-1 hash of the encoded info dictionary.
// The info dictionary is first encoded into a bencoded string, and then the SHA-1 hash is computed.
//
// Parameters:
// - parser: A pointer to the Parser instance containing the encoder.
// - info: A map[string]interface{} representing the info dictionary to be hashed.
//
// Returns:
// - A string containing the hexadecimal representation of the SHA-1 hash.
// - An error if the encoding of the info dictionary fails.
func CalculateInfoHash(parser *Parser, info map[string]interface{}) (string, error) {
	encoded, err := parser.encoder.encodeDict(info)
	if err != nil {
		return "", err
	}

	hash := sha1.Sum([]byte(encoded))
	return hex.EncodeToString(hash[:]), nil
}
