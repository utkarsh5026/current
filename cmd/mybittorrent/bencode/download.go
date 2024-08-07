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
	MsgUnChoke    = 1
	MsgInterested = 2
	MsgBitfield   = 5
	MsgRequest    = 6
	MsgPiece      = 7
)

const (
	BlockSize = 16 * 1024 // 16KB
)

// DownLoadFile downloads the specified pieces of a torrent file and writes them to an output file.
//
// Parameters:
// - t: A TorrentInfo struct containing information about the torrent.
// - outputFile: A string representing the path to the output file where the downloaded data will be written.
// - pieceIndices: A variadic integer slice representing the indices of the pieces to be downloaded.
//
// Returns:
// - An error if any step in the process fails.
func DownLoadFile(t TorrentInfo, outputFile string, pieceIndices ...int) error {
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

	if err := establishConnectionToDownloadPiece(conn); err != nil {
		return fmt.Errorf("error establishing connection: %w", err)
	}

	var fileData []byte
	for _, pieceIdx := range pieceIndices {
		piece, err := downloadPiece(conn, t, pieceIdx)
		if err != nil {
			return fmt.Errorf("error downloading piece: %w", err)
		}

		fileData = append(fileData, piece...)
	}

	return os.WriteFile(outputFile, fileData, os.ModePerm)
}

// establishConnectionToDownloadPiece establishes a connection to download a piece from a peer.
//
// Parameters:
// - conn: A net.Conn representing the TCP connection to the peer.
//
// Returns:
// - An error if any step in the process fails.
func establishConnectionToDownloadPiece(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	if err := waitForBitField(reader); err != nil {
		return err
	}

	if err := sendInterested(conn); err != nil {
		return err
	}

	if err := waitForUnChoke(reader); err != nil {
		return err
	}

	return nil
}

// downloadPiece downloads a specific piece from a peer and verifies its hash.
//
// Parameters:
// - conn: A net.Conn representing the TCP connection to the peer.
// - t: A TorrentInfo struct containing information about the torrent.
// - pieceIndex: An integer representing the index of the piece to be downloaded.
//
// Returns:
// - A byte slice containing the downloaded piece data.
// - An error if any step in the process fails.
func downloadPiece(conn net.Conn, t TorrentInfo, pieceIndex int) ([]byte, error) {
	pieceSize := t.PieceLength
	pieceCnt := int(math.Ceil(float64(t.Length) / float64(pieceSize)))
	if pieceIndex == pieceCnt-1 {
		pieceSize = t.Length % t.PieceLength
	}

	blockCnt := int(math.Ceil(float64(pieceSize) / float64(BlockSize)))

	var piece []byte
	for i := 0; i < blockCnt; i++ {
		blockLength := BlockSize
		if i == blockCnt-1 {
			blockLength = int(pieceSize) - ((blockCnt - 1) * BlockSize)
		}

		index := i * BlockSize
		// Send request for block
		if err := sendRequest(conn, pieceIndex, index, blockLength); err != nil {
			return nil, fmt.Errorf("error sending request: %w", err)
		}

		// Receive block
		block, err := receivePiece(conn, pieceIndex, index)
		if err != nil {
			return nil, fmt.Errorf("error receiving block: %w", err)
		}

		piece = append(piece, block...)
	}

	if !verifyPiece(piece, []byte(t.PieceHashes[pieceIndex])) {
		return nil, fmt.Errorf("piece hash does not match")
	}
	return piece, nil
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
	message := make([]byte, 17)
	binary.BigEndian.PutUint32(message[:4], uint32(13)) // Length of the message
	message[4] = MsgRequest                             // Request message ID
	binary.BigEndian.PutUint32(message[5:9], uint32(index))
	binary.BigEndian.PutUint32(message[9:13], uint32(begin))
	binary.BigEndian.PutUint32(message[13:17], uint32(length))

	_, err := conn.Write(message)
	return err
}

// receivePiece reads a piece message from the peer and verifies its index and begin offset.
//
// Parameters:
// - conn: A net.Conn representing the TCP connection to the peer.
// - expectedIndex: An integer representing the expected piece index.
// - expectedBegin: An integer representing the expected beginning offset within the piece.
//
// Returns:
// - A byte slice containing the piece data.
// - An error if any step in the process fails or if the received piece does not match the expected index and begin offset
func receivePiece(conn net.Conn, expectedIndex, expectedBegin int) ([]byte, error) {
	reader := bufio.NewReader(conn)
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
		return 0, 0, err
	}

	length := int(binary.BigEndian.Uint32(lengthBuf))
	if length == 0 {
		return 0, 0, nil // keep alive message
	}

	idBuffer := make([]byte, 1)
	if _, err := io.ReadFull(reader, idBuffer); err != nil {
		return 0, 0, err
	}

	return length, idBuffer[0], nil
}
