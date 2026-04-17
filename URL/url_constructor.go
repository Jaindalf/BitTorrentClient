package main

import (
	"crypto/rand"
	
	"encoding/json"
	"fmt"
	"os"
	"net/http"
)
import . "BitTorrentClient/Bdecoder"

// Info hashes


func Url(t Torrent) {
	url := ""
	url += t.Announce
	url += "?"

	fmt.Println(t.Info.InfoHash)
	url += "info_hash="
	url += HashEncoder(t.Info.InfoHash)
	url += "&peer_id="

	url += HashEncoder(generatePeerID())

	url += "&port=6881"
	url += "&uploaded=0"
	url += "&downloaded=0"
	url += "&left=0"
	url += "&compact=1"

	fmt.Println(url)

}
func generatePeerID() [20]byte {
	var peerID [20]byte

	prefix := "-PC0001-" // 8 bytes

	copy(peerID[:], prefix)

	// Fill remaining 12 bytes with random data
	_, err := rand.Read(peerID[8:])
	if err != nil {
		panic(err)
	}

	return peerID
}

func HashEncoder(ih [20]byte) string {

	result := ""

	for _, b := range ih {
		if (b >= '0' && b <= '9') ||
			(b >= 'a' && b <= 'z') ||
			(b >= 'A' && b <= 'Z') ||
			(b == '.' || b == '-' || b == '_' || b == '~') {

			result += string(b)

		} else {
			result += fmt.Sprintf("%%%02X", b)
		}
	}

	return result

}






func main() {

	data, err := os.ReadFile(`C:\Users\Pranjal\Desktop\Projects\BitTorrentClient\MX-25.1_fluxbox_x64.iso.torrent`)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	_, v, rawInfoDict := ParseValue(data, 0)
	hash := getInfoHash(rawInfoDict)
	PrettyPrint(v, 0)
	d := v.(BDict)
	t := BuildTorrent(d, hash)
	
	Url(t)

	inf, e := os.ReadFile(`C:\Users\Pranjal\Desktop\Projects\BitTorrentClient\resp`)
	if e != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	_, vs, _ := ParseValue(inf, 0)
	PrettyPrint(vs, 0)
	vw, er := json.MarshalIndent(vs, "", "  ")
	if er != nil {
		panic(er)
	}

	os.WriteFile("output.json", vw, 0644)
}
