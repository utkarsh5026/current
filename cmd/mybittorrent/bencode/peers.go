package bencode

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

// CallTracker sends a request to the tracker URL specified in the TorrentInfo and returns the response.
//
// Parameters:
// - t: A TorrentInfo struct containing the torrent metadata.
//
// Returns:
// - A byte slice containing the response from the tracker.
// - An error if any step in the process fails.
func CallTracker(t TorrentInfo) ([]byte, error) {
	infoHash, err := hex.DecodeString(t.InfoHash)
	if err != nil {
		return nil, err
	}

	peerId := make([]byte, 20)
	if _, err := rand.Read(peerId); err != nil {
		return nil, err
	}
	params := url.Values{
		"info_hash":  {string(infoHash)},
		"peer_id":    {string(peerId)},
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

// ExtractPeers extracts peer information from the tracker response.
//
// Parameters:
// - trackerResp: A byte slice containing the response from the tracker.
//
// Returns:
// - A slice of strings, each representing a peer in the format "IP:port".
// - An error if the decoding fails or if the "peers" field is missing.
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

	ipSize := 4
	portSize := 2
	for i := 0; i < len(peersData); i += ipSize + portSize {
		ip := net.IP(peersData[i : i+ipSize])
		portStart := i + ipSize
		port := binary.BigEndian.Uint16([]byte(peersData[portStart : portStart+portSize]))
		peers = append(peers, fmt.Sprintf("%s:%d", ip.String(), port))
	}

	return peers, nil
}

// HandShakeWithPeer establishes a TCP connection with a peer and performs a BitTorrent handshake.
//
// Parameters:
// - t: A TorrentInfo struct containing the torrent metadata.
// - peerAddress: A string containing the address of the peer in the format "IP:port".
//
// Returns:
// - A byte slice containing the handshake response from the peer.
// - An error if any step in the process fails
func HandShakeWithPeer(t TorrentInfo, peerAddress string) (net.Conn, []byte, error) {
	conn, err := connectToPeer(peerAddress, t)
	if err != nil {
		return nil, nil, err
	}

	response := make([]byte, 68)
	if _, err := conn.Read(response); err != nil {
		return nil, nil, err
	}

	return conn, response, nil
}

func connectToPeer(addr string, t TorrentInfo) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	infoHashBytes, err := hex.DecodeString(t.InfoHash)
	if err != nil {
		return nil, err
	}

	message := createHandShakeMessage(infoHashBytes, "00112233445566778899")
	if _, err := conn.Write(message); err != nil {
		return nil, err
	}

	return conn, nil
}

// createHandShakeMessage creates a BitTorrent handshake message.
//
// Parameters:
// - infoHash: A byte slice containing the info hash of the torrent.
// - peerId: A string containing the peer ID.
//
// Returns:
// - A byte slice representing the handshake message.
func createHandShakeMessage(infoHash []byte, peerId string) []byte {
	protocolString := "BitTorrent protocol"
	handShake := make([]byte, 0, 68)

	handShake = append(handShake, byte(len(protocolString)))
	handShake = append(handShake, []byte(protocolString)...)
	handShake = append(handShake, make([]byte, 8)...) // 8 reserved bytes (0x00)
	handShake = append(handShake, infoHash...)
	handShake = append(handShake, []byte(peerId)...)

	return handShake
}
