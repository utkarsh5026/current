package bencode

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"os"
)

const (
	MsgChoke         = 0
	MsgUnChoke       = 1
	MsgInterested    = 2
	MsgNotInterested = 3
	MsgHave          = 4
	MsgBitfield      = 5
	MsgRequest       = 6
	MsgPiece         = 7
	MsgCancel        = 8
)

const (
	BlockSize = 16 * 1024 // 16KB
)

func DownLoadFile(t TorrentInfo, outputFile string, pieceIdx int) error {
	trackerResp, err := CallTracker(t)
	if err != nil {
		return err
	}

	peers, err := ExtractPeers(trackerResp)
	if err != nil {
		return err
	}

	var conn net.Conn
	for _, peer := range peers {
		conn, _, err = HandShakeWithPeer(t, peer)
		if err == nil {
			break
		}
	}

	if conn == nil {
		return fmt.Errorf("failed to connect to any peers")
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("error closing connection")
		}
	}(conn)

	fmt.Println("Downloading piece", pieceIdx)
	piece, err := downloadPiece(conn, t, pieceIdx)
	if err != nil {
		return fmt.Errorf("error downloading piece: %w", err)
	}

	if !verifyPiece(piece, []byte(t.PieceHashes[pieceIdx])) {
		return fmt.Errorf("piece hash does not match")
	}

	return os.WriteFile(outputFile, piece, 0644)
}

func downloadPiece(conn net.Conn, t TorrentInfo, pieceIndex int) ([]byte, error) {
	reader := bufio.NewReader(conn)

	fmt.Println("Waiting for bitfield")
	if err := waitForBitField(reader); err != nil {
		return nil, err
	}

	fmt.Println("Sending interested")
	if err := sendInterested(conn); err != nil {
		return nil, err
	}

	fmt.Println("Waiting for unchoke")
	if err := waitForUnChoke(reader); err != nil {
		return nil, err
	}

	pieceSize := t.PieceLength
	pieceCnt := int(math.Ceil(float64(t.Length) / float64(pieceSize)))
	if pieceIndex == pieceCnt-1 {
		pieceSize = t.Length % t.PieceLength
	}

	blockCnt := int(math.Ceil(float64(pieceSize) / float64(BlockSize)))

	var data []byte
	for i := 0; i < blockCnt; i++ {
		blockLength := BlockSize
		if i == blockCnt-1 {
			blockLength = int(pieceSize) - ((blockCnt - 1) * BlockSize)
		}

		// Send request for block
		if err := sendRequest(conn, pieceIndex, i*BlockSize, blockLength); err != nil {
			return nil, fmt.Errorf("error sending request: %w", err)
		}

		// Receive block
		block, err := receivePiece(reader, pieceIndex, i*BlockSize)
		if err != nil {
			return nil, fmt.Errorf("error receiving block: %w", err)
		}

		data = append(data, block...)
	}

	return data, nil
}

// waitForBitField waits for a bitfield message from the peer.
//
// Parameters:
// - reader: A pointer to a bufio.Reader from which the bitfield message will be read.
//
// Returns:
// - An error if any step in the process fails.
func waitForBitField(reader *bufio.Reader) error {
	for {
		msgLen, msgId, err := readMessageHeader(reader)
		if err != nil {
			return err
		}

		if msgId == MsgBitfield { // Bitfield message
			if _, err := reader.Discard(msgLen - 1); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// sendInterested sends an "interested" message to the peer over the given TCP connection.
//
// Parameters:
// - conn: A net.Conn representing the TCP connection to the peer.
//
// Returns:
// - An error if the message could not be sent.
func sendInterested(conn net.Conn) error {
	message := make([]byte, 5)
	binary.BigEndian.PutUint32(message[:4], 1) // Length of the message
	message[4] = MsgInterested                 // Interested message ID

	if _, err := conn.Write(message); err != nil {
		return err
	}
	return nil
}

// waitForUnChoke waits for an "unchoke" message from the peer.
//
// Parameters:
// - reader: A pointer to a bufio.Reader from which the "unchoke" message will be read.
//
// Returns:
// - An error if any step in the process fails or if the connection is closed unexpectedly.
func waitForUnChoke(reader *bufio.Reader) error {
	for {
		_, msgId, err := readMessageHeader(reader)
		if err != nil {
			return err
		}

		if msgId == MsgUnChoke {
			break
		}
	}
	return nil
}

// sendRequest sends a request message to the peer over the given TCP connection.
//
// Parameters:
// - conn: A net.Conn representing the TCP connection to the peer.
// - index: An integer representing the piece index being requested.
// - begin: An integer representing the beginning offset within the piece.
// - length: An integer representing the length of the block being requested.
//
// Returns:
// - An error if the message could not be sent.
func sendRequest(conn net.Conn, index, begin, length int) error {
	fmt.Println("sending request", index, begin, length)
	message := make([]byte, 17)
	binary.BigEndian.PutUint32(message[:4], uint32(13)) // Length of the message
	message[4] = MsgRequest                             // Request message ID
	binary.BigEndian.PutUint32(message[5:9], uint32(index))
	binary.BigEndian.PutUint32(message[9:13], uint32(begin))
	binary.BigEndian.PutUint32(message[13:17], uint32(length))

	_, err := conn.Write(message)
	return err
}

func receivePiece(reader *bufio.Reader, expectedIndex, expectedBegin int) ([]byte, error) {
	msgLen, msgId, err := readMessageHeader(reader)
	if err != nil {
		return nil, err
	}

	if msgId != MsgPiece {
		return nil, fmt.Errorf("expected piece message but got %d", msgId)
	}

	payload := make([]byte, msgLen-1)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}

	index := int(binary.BigEndian.Uint32(payload[:4]))
	begin := int(binary.BigEndian.Uint32(payload[4:8]))

	if index != expectedIndex || begin != expectedBegin {
		return nil, fmt.Errorf("expected block from index %d begin %d but got index %d begin %d", expectedIndex, expectedBegin, index, begin)
	}

	return payload[8:], nil
}

// readMessageHeader reads the message header from the given buffered reader.
//
// Parameters:
// - reader: A pointer to a bufio.Reader from which the message header will be read.
//
// Returns:
// - An integer representing the length of the message.
// - A byte representing the message ID.
// - An error if any step in the process fails.
func readMessageHeader(reader *bufio.Reader) (int, byte, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(reader, lengthBuf); err != nil {
		fmt.Println("error reading length")
		return 0, 0, err
	}

	length := int(binary.BigEndian.Uint32(lengthBuf))
	if length == 0 {
		return 0, 0, nil // keep alive message
	}

	idBuffer := make([]byte, 1)
	if _, err := io.ReadFull(reader, idBuffer); err != nil {
		fmt.Println("error reading id")
		return 0, 0, err
	}

	return length, idBuffer[0], nil
}
