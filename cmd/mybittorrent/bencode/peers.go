package bencode

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

func CallTracker(t TorrentInfo) ([]byte, error) {
	infoHash, err := hex.DecodeString(t.InfoHash)

	if err != nil {
		return nil, err
	}

	params := url.Values{
		"info_hash":  {string(infoHash)},
		"peer_id":    {"00112233445566778899"},
		"port":       {"6881"},
		"uploaded":   {"0"},
		"downloaded": {"0"},
		"left":       {fmt.Sprintf("%d", t.Length)},
		"compact":    {"1"},
	}

	reqUrl := fmt.Sprintf("%s?%s", t.Announce, params.Encode())
	resp, err := http.Get(reqUrl)

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	return io.ReadAll(resp.Body)
}

func ExtractPeers(trackerResp []byte) ([]string, error) {
	decoder := NewDecoder(string(trackerResp))
	decoded, err := decoder.Decode()

	if err != nil {
		return nil, err
	}

	dict := decoded.(map[string]interface{})
	peersData, ok := dict["peers"].(string)

	if !ok {
		return nil, fmt.Errorf("missing peers")
	}

	var peers []string

	for i := 0; i < len(peersData); i += 6 {
		ip := net.IP(peersData[i : i+4])
		port := binary.BigEndian.Uint16([]byte(peersData[i+4 : i+6]))
		peers = append(peers, fmt.Sprintf("%s:%d", ip.String(), port))
	}

	return peers, nil
}
