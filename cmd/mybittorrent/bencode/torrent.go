package bencode

import "fmt"

type TorrentInfo struct {
	Announce string
	Length   int64
	Info     map[string]interface{}
}

func (t TorrentInfo) PrintStats() {
	fmt.Printf("Tracker URL: %v\n", t.Announce)
	fmt.Printf("Length: %v\n", t.Length)
}
