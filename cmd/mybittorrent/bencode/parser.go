package bencode

import (
	"fmt"
)

type Parser struct {
	decoder *Decoder
}

func CreateParser(input string) *Parser {
	return &Parser{decoder: NewDecoder(input)}
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

	return &TorrentInfo{
		Announce: announce,
		Length:   int64(length),
		Info:     info,
	}, nil
}
