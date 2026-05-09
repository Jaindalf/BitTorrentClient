package main

import (
	"bytes"
	"encoding/binary"
	//"flag"

	//"encoding/gob"
	"fmt"
	"io"
	"net"

	//	"os"
	"strconv"
	"strings"
	"time"

	//"os"
	. "BitTorrentClient/Bdecoder"
	"math"

	. "BitTorrentClient/URL"
	"sync"
)

//var h []byte

const pstr = "BitTorrent protocol"
const pstrlen = byte(len(pstr))

// handshake: <pstrlen><pstr><reserved><info_hash><peer_id>
type Handshake struct {
	pstrlen  int
	pstr     string
	infoHash [20]byte
	peerID   [20]byte
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

func NormalizeIP(ip string) string {
	if strings.HasPrefix(ip, "::ffff:") {
		return strings.TrimPrefix(ip, "::ffff:")
	}
	return ip
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
	return buf

}

func VerifyHandshake(buf []byte, t Torrent) bool {
	offset := 0
	if len(buf) < 68 {
		fmt.Println("Incomplete hanshake response")
		return false

	}
	pstrlen := int(buf[offset])

	if pstrlen != 19 {
		fmt.Println("invalid pstrlen: ", pstrlen)
		return false
	}
	offset++

	pstr := string(buf[offset : offset+pstrlen])
	if pstr != "BitTorrent protocol" {
		fmt.Println("invalid protocol: ", pstr)
		return false
	}

	offset += len(pstr)

	// skip reserved (8 bytes)
	offset += 8

	var infoHash [20]byte
	copy(infoHash[:], buf[offset:offset+20])

	if !bytes.Equal(t.Info.InfoHash[:], infoHash[:]) {
		fmt.Println("Invalid Info Hash:")
		return false

	}
	offset += 20
	return true

}

type Connection struct {
	conn     net.Conn
	choked   bool
	bitfield []byte
	PeerIp   string
}

func AttemptConn(t Torrent, resp TrackerResponse, ti time.Duration) []Connection {

	var Connections []Connection

	for index, val := range resp.Peers {
		ip := NormalizeIP(val.IP)
		addr := net.JoinHostPort(ip, strconv.Itoa(val.Port))

		conn, err := net.DialTimeout("tcp", addr, ti)
		if err != nil {
			fmt.Println("Missed index\t", index)
			continue
		}
		fmt.Println("Connected[", index, "] to: ", addr)
		h := BuildHandshake(t)
		_, e := conn.Write(h)
		if e != nil {
			fmt.Println("Error:", e)
			continue
		}

		buffer := make([]byte, 68)

		_, er := io.ReadFull(conn, buffer)
		if er != nil {
			fmt.Println("Error:", err)
			continue
		}

		if !VerifyHandshake(buffer, t) {
			//return nil
			continue
		}
		var p Connection
		p.conn = conn
		p.PeerIp = ip
		p.choked = true
		p.bitfield = nil
		Connections = append(Connections, p)
		//return conn

	}

	return Connections

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
		return &Message{ID: 100, Payload: nil} //this is not an actual id this is just so we can have a msg to return

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

func R(conn net.Conn) {

	i := 8
	for i < 100 {
		fmt.Println("READING")

		m := ReceiveMessage(conn)
		switch m.ID {

		case 5:
			fmt.Println("Bitfield")
			time.Sleep(time.Second)

		case 1:
			fmt.Println("Unchoked")
			time.Sleep(time.Second)

		case 100:
			fmt.Println("Keep-alive")
			time.Sleep(time.Second)

		}

	}

}

func S(conn net.Conn) {
	fmt.Println("SENDING")
	m := (StaticPeerMessages["Interested"])
	SendMessage(conn, &m)
	time.Sleep(time.Second)

}

func Pieces(conn net.Conn, t Torrent) {

	pieceCount := uint32(len(t.Info.Pieces) / 20)
	fmt.Println("Number of pieces:", pieceCount)

	pieceSize := uint32(t.Info.PieceLength)
	fmt.Println("Size of pieces", pieceSize)

	fileSize := pieceCount * pieceSize
	fmt.Println("Size of file(in bytes):", fileSize)
	blockSize := uint32(math.Pow(2, 14))
	var j uint32
	var i uint32

	for i = 0; i < pieceCount; i++ {

		for j = 0; j < pieceSize-1; j++ {

			offset := blockSize * j

			r := Request(i, (offset), (blockSize))
			fmt.Println("Asking for piece ", j)
			SendMessage(conn, r)
			time.Sleep(time.Second * 3)

			msg := ReceiveMessage(conn)

			fmt.Println("MSSSG ID:", msg.ID, "Response ", j)

		}

	}

}

func getBitfield(peer *Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	tries := 0
	maxtries := 3
	for tries < maxtries {
		m := ReceiveMessage(peer.conn)
		if m.ID == 5 {
			peer.bitfield = m.Payload
			fmt.Println("Received bitflied of", peer.PeerIp)
			//wg.Done()

			return

		} else {
			tries++
			time.Sleep(2 * time.Second)

		}

	}
	fmt.Println("tried for:", tries)
	//	wg.Done()

}

func unChoked(peer *Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	tries := 0
	maxtries := 3
	for tries < maxtries {
		m := ReceiveMessage(peer.conn)
		if m.ID == 1 {
			peer.choked = false
			fmt.Println("Unchoked:", peer.PeerIp)
			//wg.Done()

			return

		} else {
			tries++
			time.Sleep(2 * time.Second)

		}

	}
	fmt.Println("tried for:", tries)
	//	wg.Done()

}

func Formality(peers []Connection) {

	var wg sync.WaitGroup

	for index, val := range peers {
		fmt.Println("Index: ", index, "IpAddr:", val.PeerIp)
		wg.Add(1)
		go getBitfield(&peers[index], &wg)
	}

	wg.Wait()

	for index, val := range peers {
		i := StaticPeerMessages["Interested"]
		e := SendMessage(val.conn, &i)
		if e != nil {
			fmt.Println("[", index, "] Error sending interested msg:", e)

		}
	}

	for index, val := range peers {
		fmt.Println("Index: ", index, "IpAddr:", val.PeerIp)
		wg.Add(1)
		go unChoked(&peers[index], &wg)
	}

	wg.Wait()

	for index, val := range peers {
		fmt.Println("Index: ", index, "choked:", val.choked)

	}

	//Remove all the choked=true peers

	for index, val := range peers {
		if val.choked {

			peers = append(peers[:index], peers[index+1:]...)

		}

	}

	for index, val := range peers {
		fmt.Println("Index: ", index, "choked:", val.choked)

	}

}

func getPieceCount(t Torrent) uint32 {
	return uint32(len(t.Info.Pieces) / 20)

}
func main() {
	t := `C:\Users\Pranjal\Downloads\ubuntu-26.04-desktop-amd64.iso.torrent`
	rawRespDict, tor := GetTrackerResponse(t)
	respDict := ParseTrackerResponse(rawRespDict)

	peers := AttemptConn(tor, respDict, 3*time.Second)
	fmt.Println("No. of peers connected:", len(peers))

	Formality(peers)

	pieceCount := getPieceCount(tor)
	fmt.Println("Number of pieces:", pieceCount)

	pieceSize := uint32(tor.Info.PieceLength)
	fmt.Println("Size of pieces", pieceSize)

	fileSize := pieceCount * pieceSize
	fmt.Println("Size of file(in bytes):", fileSize)
	blockSize := uint32(math.Pow(2, 14))
	var j uint32
	var i uint32

	//fmt.Println()

	//Pieces(con, tor)
}
