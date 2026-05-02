package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
	//"os"
	"math"
	. "BitTorrentClient/Bdecoder"

	. "BitTorrentClient/URL"
)

const pstr = "BitTorrent protocol"
const pstrlen = byte(len(pstr))

//handshake: <pstrlen><pstr><reserved><info_hash><peer_id>

func VerifyHandshake(buf []byte, t Torrent, pid []byte) {
	offset := 0
	if len(buf) < 68 {
		panic("Incomplete hanshake response")

	}
	pstrlen := int(buf[offset])

	if pstrlen != 19 {
		fmt.Println("invalid pstrlen: ", pstrlen)
	}
	offset++

	pstr := string(buf[offset : offset+pstrlen])
	if pstr != "BitTorrent protocol" {
		fmt.Println("invalid protocol: ", pstr)
	}

	offset += len(pstr)

	// skip reserved (8 bytes)
	offset += 8

	var infoHash [20]byte
	copy(infoHash[:], buf[offset:offset+20])

	if !bytes.Equal(t.Info.InfoHash[:], infoHash[:]) {
		fmt.Println("Invalid Info Hash:")

	}
	offset += 20

	var peerID [20]byte
	copy(peerID[:], buf[offset:offset+20])

	if !bytes.Equal(pid[:], peerID[:]) {
		fmt.Println("Invalid peer ID:")

		fmt.Println("----PID-----", pid)
		fmt.Println("----peerID----", peerID)
	}

}

type Handshake struct {
	pstrlen  int
	pstr     string
	infoHash [20]byte
	peerID   [20]byte
}

func BuildHandshake(t Torrent) []byte {

	buf := make([]byte, (49 + len(pstr)))
	writer := 0

	buf[writer] = pstrlen //This is 1 byte
	writer++
	copy(buf[writer:], pstr)
	writer += len(pstr)

	for i := 0; i < 8; i++ {
		buf[writer+i] = 0

	}
	writer += 8

	copy(buf[writer:], t.Info.InfoHash[:])
	writer += 20 //Info Hash is 20 bytes

	copy(buf[writer:], PeerId[:])
	writer += 20
	fmt.Println(buf)
	return buf

}

func NormalizeIP(ip string) string {
	if strings.HasPrefix(ip, "::ffff:") {
		return strings.TrimPrefix(ip, "::ffff:")
	}
	return ip
}

var h []byte

func AttemptConn(t Torrent, resp TrackerResponse, ti time.Duration) net.Conn {

	hit := 0
	miss := 0

	for index, val := range resp.Peers {
		ip := NormalizeIP(val.IP)
		addr := net.JoinHostPort(ip, strconv.Itoa(val.Port))

		conn, err := net.DialTimeout("tcp", addr, ti)
		if err != nil {
			miss++
			fmt.Println("Missed index\t", index)
			continue
		}
		hit++
		fmt.Println("Connected\t", addr)
		//defer conn.Close()
		h := BuildHandshake(t)
		_, e := conn.Write(h)
		if e != nil {
			fmt.Println("Error:", e)
		}

		buffer := make([]byte, 68)

		_, er := io.ReadFull(conn, buffer)
		if er != nil {
			fmt.Println("Error:", err)
		}

		fmt.Println("val pid", (val.ID))
		VerifyHandshake(buffer, t, []byte(val.ID))
		return conn

	}

	fmt.Println("\nSummary:")

	fmt.Println("Hits:", hit)
	fmt.Println("Misses:", miss)
	return nil

}

type Message struct {
	ID      uint8
	Payload []byte
}

var StaticPeerMessages = map[string]Message{
	"Choke":        {ID: 0, Payload: nil},
	"UnChoke":      {ID: 1, Payload: nil},
	"Interested":   {ID: 2, Payload: nil},
	"UnInterested": {ID: 3, Payload: nil},
}

func KeepAlive(conn net.Conn) error {
	buf := make([]byte, 4)
	_, er := conn.Write(buf)
	return er

}

func Have(pieceIndex uint32) *Message {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:4], pieceIndex)
	return &Message{ID: 4, Payload: buf}
}

func BitField(bitfield []byte) *Message {
	return &Message{ID: 5, Payload: bitfield}

}

func Request(index, begin, length uint32) *Message {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], index)
	binary.BigEndian.PutUint32(buf[4:8], begin)
	binary.BigEndian.PutUint32(buf[8:12], length)
	return &Message{ID: 6, Payload: buf}
}

func Cancel(index, begin, length uint32) *Message {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], index)
	binary.BigEndian.PutUint32(buf[4:8], begin)
	binary.BigEndian.PutUint32(buf[8:12], length)
	return &Message{ID: 8, Payload: buf}
}

func Piece(index, begin uint32, block []byte) *Message {
	buf := make([]byte, (8 + len(block)))
	binary.BigEndian.PutUint32(buf[0:4], index)
	binary.BigEndian.PutUint32(buf[4:8], begin)
	copy(buf[8:], block[:])
	return &Message{ID: 7, Payload: buf}
}

func Port(listenPort uint16) *Message {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf[0:2], listenPort)
	return &Message{ID: 9, Payload: buf}

}

func SendMessage(conn net.Conn, msg *Message) error {
	//First filter out keep-alive

	length := uint32(len(msg.Payload) + 1) //the 1 is for the msg.ID
	buf := make([]byte, (length + 4))      //the 4 is for the message length

	//fill in the len field
	binary.BigEndian.PutUint32(buf[0:4], length)
	//fill in msg.Id
	buf[4] = msg.ID

	// payload
	copy(buf[5:], msg.Payload)

	_, err := conn.Write(buf)
	fmt.Println(err)
	return err

}

func ReceiveMessage(conn net.Conn) *Message {

	lenbuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lenbuf)
	if err != nil {
		fmt.Println("Error reading Length:", err)
	}

	length := binary.BigEndian.Uint32(lenbuf)
	if length == 0 {
		fmt.Println("KEEP-ALIVE MESSAGE")
		return &Message{ID:100,Payload: nil}

	}

	fmt.Println("LENGTH:", length)

	msgBuf := make([]byte, length)
	_, err = io.ReadFull(conn, msgBuf)
	if err != nil {
		fmt.Println("failed to read message: %w", err)
		return nil
	}

	m := Message{ID: uint8(msgBuf[0]), Payload: msgBuf[1:]}

	fmt.Println("msg id:", m.ID)

	if length != uint32((len(m.Payload) + 1)) {
		fmt.Println("Length missmatch.Incomplete payload.")

	}
	return &m

}

func Pieces(conn net.Conn,t Torrent) {

	pieceCount := uint32(len(t.Info.Pieces) / 20)
	fmt.Println("Number of pieces:", pieceCount)

	pieceSize := uint32(t.Info.PieceLength)
	fmt.Println("Size of pieces", pieceSize)

	fileSize:=pieceCount*pieceSize
	fmt.Println("Size of file(in bytes):",fileSize)
	blockSize:=uint32(math.Pow(2,14))
	var  j uint32;
	var i uint32;

	for  i=0;i<pieceCount;i++{

		for j=0;j<pieceSize-1;j++{

			offset:=blockSize*j

			r:=Request(i,(offset),(blockSize))
			fmt.Println("Asking for piece ",j)
			SendMessage(conn,r)
			time.Sleep(time.Second*3)

			msg:=ReceiveMessage(conn)
			
			fmt.Println("MSSSG ID:",msg.ID,"Response ",j)

		}

	}

}

/*func main() {
	b := `C:\Users\Pranjal\Downloads\ubuntu-26.04-desktop-amd64.iso.torrent`

	data, err := os.ReadFile(b)
	if err != nil {
		fmt.Println("Error reading the torrent file.")
		
	}


	_, rawDict, rawInfoDict := ParseValue(data, 0)
	infoDictHash := GetInfoHash(rawInfoDict)
	Dict := rawDict.(BDict)
	fileTorrent := BuildTorrent(Dict, infoDictHash)
	Pieces(fileTorrent)
}*/

func main() {
	b := `C:\Users\Pranjal\Downloads\ubuntu-26.04-desktop-amd64.iso.torrent`
	//b:=`C:\Users\Pranjal\Downloads\bazzite-43.20260420-deck-stable-amd64.iso.torrent`
	//s:=`one-piece.torrent`
	//b := `C:\Users\Pranjal\Downloads\xubuntu-26.04-desktop-amd64.iso.torrent`
	//b:=`C:\Users\Pranjal\Downloads\cosmos-laundromat.torrent`
	rawResp, torrent := GetTrackerResponse(b)
	resp := ParseTrackerResponse(rawResp)
	for i, v := range resp.Peers {
		fmt.Println("Index:", i, "\tID:", v.ID, "\tIp:", v.IP, "\tPort:", v.Port)
	}

	h = BuildHandshake(torrent)

	con := AttemptConn(torrent, resp, time.Second)

	msg := ReceiveMessage(con)
	if msg.ID == 5 {
		fmt.Println("Bitfield")
		fmt.Printf("Received bitfield (%d bytes)\n", len(msg.Payload))
		i:=(StaticPeerMessages["Interested"])
		SendMessage(con,&i)

		m:=ReceiveMessage(con)
		if m.ID==1{
			    fmt.Println("Unchoked!",m.ID)

		}
	}
	//e := KeepAlive(con)
	//fmt.Println(e)
	m := StaticPeerMessages["Interested"]
	SendMessage(con, &m)

	Pieces(con,torrent)


}

