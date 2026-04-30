package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	. "BitTorrentClient/Bdecoder"

	. "BitTorrentClient/URL"
)

const pstr = "BitTorrent protocol"
const pstrlen = byte(len(pstr))

//handshake: <pstrlen><pstr><reserved><info_hash><peer_id>

func VerifyHandshake(buf []byte, t Torrent,pid []byte) {
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

	if (!bytes.Equal(t.Info.InfoHash[:],infoHash[:])){
		fmt.Println("Invalid Info Hash:")
		
	}
	offset += 20

	var peerID [20]byte
	copy(peerID[:], buf[offset:offset+20])

	if (!bytes.Equal(pid[:],peerID[:])){
		fmt.Println("Invalid peer ID:")

		fmt.Println("----PID-----",pid)
		fmt.Println("----peerID----",peerID)
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

func AttemptConn(t Torrent, resp TrackerResponse, ti time.Duration) {

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
		defer conn.Close()
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

		fmt.Println("val pid",(val.ID))
		VerifyHandshake(buffer,t,[]byte(val.ID))
		return

	}

	fmt.Println("\nSummary:")

	fmt.Println("Hits:", hit)
	fmt.Println("Misses:", miss)

}


func main() {
	g := `C:\Users\Pranjal\Downloads\ubuntu-26.04-desktop-amd64.iso.torrent`
	//s:=`one-piece.torrent`
	//t := `MX-25.1_fluxbox_x64.iso.torrent`
	rawResp, torrent := GetTrackerResponse(g)
	resp := ParseTrackerResponse(rawResp)
	for i, v := range resp.Peers {
		fmt.Println("Index:", i,"\tID:", v.ID,"\tIp:", v.IP, "\tPort:", v.Port)
	}

	h = BuildHandshake(torrent)
	//fmt.Println("ip of 692:", resp.Peers[3].Port)
	//port:=string(resp.Peers[3].Port)
	//	fmt.Println(port)

	AttemptConn(torrent,resp, time.Second)

}
