package main

import (
	. "BitTorrentClient/Bdecoder"
	. "BitTorrentClient/Handshake"
	. "BitTorrentClient/URL"
	"crypto/sha1"
	"fmt"
	"sync"

	//"math"
	//"os"
	"time"
)
type PieceStruct struct {
	PieceIndex int
	Data []byte

}
func CreateWorkQueue(tor Torrent) []byte {
	pieces := GetPieceCount(tor)
	var  workQueue []byte;

	if pieces%8 == 0 {
		workQueue = make([]byte, pieces/8)
		

	} else {
		workQueue = make([]byte, (pieces/8)+1)

	}

	for i := range workQueue {
		workQueue[i] = 0x00
	}
	return workQueue

	

}

func GetWork(peer Connection,workQueue []byte){

	

	


}


func CheckIntegrity(piece PieceStruct,t *Torrent ) bool{

	ph:=sha1.Sum(piece.Data)
	th:=[20]byte(t.Info.Pieces[20*piece.PieceIndex:(20*piece.PieceIndex+20)])

	if  th==ph{
		return true

	}else{
		return false
	}

	
}

func WriteToDisk( piece PieceStruct)// what is the size of each piece (offset)

func main() {
	t := `C:\Users\Pranjal\Downloads\ubuntu-26.04-desktop-amd64.iso.torrent`
	tor:=GetTorrent(t)
	trackerResp:=GetTrackerResponse(tor)

	peers := AttemptConn(tor, trackerResp, 3*time.Second)
	fmt.Println("No. of peers connected:", len(peers))
	
	

	Formality(peers)

	pieceCount := GetPieceCount(tor)
	fmt.Println("Number of pieces:", pieceCount)

	pieceSize := uint32(tor.Info.PieceLength)
	fmt.Println("Size of pieces", pieceSize)

	fileSize := pieceCount * pieceSize
	fmt.Println("Size of file(in bytes):", fileSize)
	//blockSize := uint32(math.Pow(2, 14))
	wq:=CreateWorkQueue(tor)
	fmt.Println(wq)
	//firstPiece := DownloadPiece(peers[0], 0, pieceSize, blockSize)



	
	
}

