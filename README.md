# 🌊 BitTorrent Client Implementation

Hey! This is my BitTorrent client built during the [Codecrafters](https://codecrafters.io/) challenge. 


[![progress-banner](https://backend.codecrafters.io/progress/bittorrent/9bfe2f08-b35d-49aa-86d3-32c54792e640)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)


## 🚀 Features

- ✨ Parse `.torrent` files and extract metadata
- 🤝 Connect to peers using the BitTorrent protocol
- 📡 Communicate with trackers to discover peers
- 📦 Download pieces from multiple peers simultaneously
- ✅ Verify downloaded pieces using SHA1 hashing
- 📊 Basic download progress tracking


## 🛠️ Technical Implementation

This client implements core BitTorrent protocol features including:

- **Bencode Parser**: Custom implementation for encoding/decoding bencode format
- **Peer Wire Protocol**: Handles peer communication and piece exchange
- **Tracker Communication**: Manages tracker requests and peer discovery
- **Piece Management**: Downloads and verifies file pieces
- **TCP Connections**: Handles concurrent peer connections


### Planned Features
- 🚄 Multi-threaded downloading for improved performance
- 📊 Real-time download progress visualization
- 🔄 Resume interrupted downloads
- 🎯 Selective file downloading from multi-file torrents
- 🔒 Support for encrypted peer connections
- 📱 Web UI for remote management
- 💾 Configurable download queue management
- 🌡️ Bandwidth throttling and scheduling
- 🔍 DHT (Distributed Hash Table) support for trackerless torrents


## 🙏 Big Thanks To

- The awesome folks at [Codecrafters](https://codecrafters.io/)
- The [BitTorrent Protocol Spec](https://www.bittorrent.org/beps/bep_0003.html)
- [Kristen Widman's super helpful guide](https://blog.jse.li/posts/torrent/)