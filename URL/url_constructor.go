package main
import(
	
	"os"
	"fmt"
	"crypto/sha1"
	"crypto/rand"
)
import . "BitTorrentClient/Bdecoder"


//Info hashes 
func infohash( ih []byte) [20]byte{
	hash:=sha1.Sum(ih)
	return hash
}

func Url(t Torrent){
	url:=""
	url+=t.Announce
	url+="?"
	
	fmt.Println(t.Info.InfoHash)
	url+="info_hash="
	url+=HashEncoder(t.Info.InfoHash)
	url+="&peer_id="

	url+=HashEncoder(generatePeerID())

	url+="&port=6881"
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


func  HashEncoder(ih [20]byte) string{

	result:=""
	
	for _,b:=range ih{
		if (b>='0' && b<='9')||
		 	(b>='a' && b<='z')||
			(b>='A' && b<='Z')||
			 (b == '.' || b == '-' || b == '_' || b == '~'){

				result+=string(b)
			
		

		}else{
			result+=fmt.Sprintf("%%%02X",b)
		}
	}

	return result


}

func main() {
	
	//archlinux-2026.04.01-x86_64.iso.torrent
	data, err := os.ReadFile(`C:\Users\Pranjal\Desktop\Projects\BitTorrentClient\MX-25.1_fluxbox_x64.iso.torrent`)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	//content := string(data)
	//fmt.Println(content)
	_, v ,ih:= ParseValue(data, 0)
	//i:=[20]byte(ih)
	hash:=infohash(ih)
	PrettyPrint(v, 0)
	d:=v.(BDict)
	t:=BuildTorrent(d,hash)
	//fmt.Println(hash)

//size:=len(t.Info.Pieces)

//for i := 0; i < size; i=i+20 {

	//fmt.Printf("Info Pieces hash: %x\n",t.Info.Pieces[i:i+20])
	
//}

Url(t)

inf, e := os.ReadFile(`C:\Users\Pranjal\Desktop\Projects\BitTorrentClient\resp`)
	if e != nil {
		fmt.Println("Error reading file:", err)
		return
	}


_,vs,_:=ParseValue(inf,0)
PrettyPrint(vs,0)
}