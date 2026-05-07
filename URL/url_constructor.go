package response

import (
	"crypto/rand"


	"fmt"

	"io"
	"net/http"
	"os"
	"encoding/gob"
)
import . "BitTorrentClient/Bdecoder"

// Info hashes
type TrackerResponse struct {
	Interval int
	Peers    []Peer
}

type Peer struct {
	ID   string
	IP   string
	Port int
}

func Url(t Torrent) string {
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
	return url

}

var PeerId [20]byte

func generatePeerID() [20]byte {
	var peerID [20]byte

	prefix := "-PC0001-" // 8 bytes

	copy(peerID[:], prefix)

	// Fill remaining 12 bytes with random data
	_, err := rand.Read(peerID[8:])
	if err != nil {
		panic(err)
	}
	PeerId=peerID

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

func GetTrackerResponse(torrentFile string) (BDict,Torrent){
	data, err := os.ReadFile(torrentFile)
	if err != nil {
		fmt.Println("Error reading the torrent file.")
		return nil,Torrent{}
	}

	_, rawDict, rawInfoDict := ParseValue(data, 0)
	infoDictHash := GetInfoHash(rawInfoDict)
	Dict := rawDict.(BDict)
	fileTorrent := BuildTorrent(Dict, infoDictHash)

	resp, e := http.Get(Url(fileTorrent))
	if e != nil {
		fmt.Println("Error in response:", e)
		return nil,Torrent{}
	}

	defer resp.Body.Close()

	rawRespBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body: ", err)
	}

	if rawRespBody==nil {
		fmt.Println("SIZE OF RAW_RESP:",len(rawRespBody))
		
	}



	_, respBody, _ := ParseValue(rawRespBody, 0)
	respBodyDict := respBody.(BDict)

	fileR,_:=os.Create("Response.gob")
	fileT,_:=os.Create("Torrent.gob")


	encoderR := gob.NewEncoder(fileR)
	encoderT := gob.NewEncoder(fileT)

	if err := encoderR.Encode(respBodyDict); err != nil {
		fmt.Println("Error encoding Response:", err)
	}
	fileR.Close()

	if err := encoderT.Encode(fileTorrent); err != nil {
		fmt.Println("Error encoding Torrent:", err)
	}
	fileT.Close()


	return respBodyDict,fileTorrent
	

}

func ParseTrackerResponse(resp BDict) TrackerResponse {
	var result TrackerResponse

	for _, item := range resp {
		switch item.Key {
		case "interval":
			result.Interval = (Get(resp, "interval")).(int)

		case "peers":
			peersRaw:=(Get(resp,"peers")).([]interface{})

			for _,p:=range peersRaw{
				//Each peer is a BDict
				peerDict:=p.(BDict)

				var peer Peer
				for _,field:=range peerDict{
					switch field.Key{
					case "id":
						peer.ID=string((Get(peerDict,"id")).([]byte))

					case "ip":
						peer.IP=string((Get(peerDict,"ip")).([]byte))
					
					case "port":
						peer.Port=(Get(peerDict,"port")).(int)	
					
					}

				}
				result.Peers=append(result.Peers, peer)


			}

		}

	}









	return result

}


