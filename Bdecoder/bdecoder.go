package main

import (
	"fmt"
	"os"
)

// BEntry represents a single entry of a bencoded dictionary.
type BEntry struct {
	Key   string      //In bencoding dictionary keys are always strings.
	Value interface{} //Interface is used because dict values can be of any type(int,string,list,dict)
}

// BDict is a dynamic array(technical term is slice) of BEntry (We have to do this to preserve the order of dict entries as go maps are unordered. )
type BDict []BEntry

type InfoDict struct {
	Name        string
	Length      int
	PieceLength int
	Pieces      []byte
}

type Torrent struct {
	Announce string
	Info     InfoDict
}

// Tells us wheter a byte is a ascii number or not.
func IsNumeric(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// Returns a word as an array of bytes and the index of the character just after the word.
func ParseString(EncodedString []byte, i int) (int, []byte) {

	bytesToRead := 0

	// Parse number before colon
	for i < len(EncodedString) && IsNumeric(EncodedString[i]) {
		digit := int(EncodedString[i] - '0')
		bytesToRead = bytesToRead*10 + digit
		i++
	}

	// Expect colon
	if i >= len(EncodedString) || EncodedString[i] != ':' {
		panic("invalid format: missing colon")
	}
	i++ // skip colon

	//fmt.Println("Length:", bytesToRead)

	if i+bytesToRead > len(EncodedString) {
		panic("invalid format: not enough bytes")
	}

	// Read exact number of bytes

	word := EncodedString[i : i+bytesToRead]
	//fmt.Println("Word:", word)
	//fmt.Println("The function returns to point", (i + bytesToRead))

	return (i + bytesToRead), word

}

// returns a number as an integr and the index of the character just after the number.
func ParseInt(EncodedInt []byte, i int) (int, int) {

	//Bounds check
	if i >= len(EncodedInt) || EncodedInt[i] != 'i' {
		panic("Invalid format:missing 'i'")
	}
	i++

	if i >= len(EncodedInt) {
		panic("Unexpected end after 'i'")
	}

	integer := 0
	digitFound := false
	sign := 1

	if len(EncodedInt) < 3 {

		panic("EncodedInt is too short.")

	}

	//Check for negative

	if EncodedInt[i] == '-' {
		sign = -1
		i++

		if i >= len(EncodedInt) {
			panic("Unexpected end after '-'")
		}

		//-0 is not allowed
		if i < len(EncodedInt) && EncodedInt[i] == '0' {
			panic("i-0e is not valid bencode")
		}

	}

	//check for leading zeroes
	//li0ee is allowed
	//li03e is not
	if EncodedInt[i] == '0' && i+1 < len(EncodedInt) && EncodedInt[i+1] != 'e' {

		panic("Corrupted data : Leading zeroes")
	}

	for i < len(EncodedInt) && EncodedInt[i] != 'e' {
		if !IsNumeric(EncodedInt[i]) {

			panic("Unexpected value in EncodedInt")
		}
		digitFound = true
		digit := int(EncodedInt[i] - '0')
		integer = integer*10 + digit
		i++

	}

	if i >= len(EncodedInt) || EncodedInt[i] != 'e' {
		panic("e was not found at the end of encoded int.")
	}

	integer = sign * integer

	if !digitFound {
		panic("Invalid integer: No digits")
	}

	//fmt.Println("int:", integer)
	//fmt.Println("index:", i) //This is the index where the limiting 'e' is present.
	return (i + 1), integer // we should return the index after the limiting e(when working with lists)
}

func ParseList(EncodedList []byte, i int) (int, []interface{}) {

	//Create a slice of interfaces
	var mySlice []interface{}

	//Check  if this is a list at all
	if EncodedList[i] != 'l' {
		panic("This is not a list.")
	}
	i++

	for i < len(EncodedList) && EncodedList[i] != 'e' {

		//ch := EncodedList[i]
		var val interface{}
		i, val = ParseValue(EncodedList, i)
		mySlice = append(mySlice, val)

	}

	if i >= len(EncodedList) || EncodedList[i] != 'e' {
		panic("List not terminated properly")
	}
	return i + 1, mySlice

}

func ParseDict(EncodedDict []byte, i int) (int, BDict) {

	//Dictionaries are encoded as follows: d<bencoded string><bencoded element>e
	// the key can only be a string

	//check if this is even a dictionary
	if EncodedDict[i] != 'd' {

		panic("This is not a dictionary.")

	}
	i++

	var dict BDict
	var prevKey string

	//Now we loop over the entire dictionary
	for i < len(EncodedDict) {

		// End of dictionary
		if EncodedDict[i] == 'e' {
			fmt.Println("End of dict reached.")
			return i + 1, dict
		}

		// Parse key (always a string)
		var key []byte
		i, key = ParseString(EncodedDict, i)
		//fmt.Println("KEY:", key)
		currentKey := string(key)

		if prevKey != "" && currentKey < prevKey {
			fmt.Println("Dictionary keys not sorted lexicographically")
		}

		prevKey = currentKey

		var value interface{}
		//ch := EncodedDict[i]

		i, value = ParseValue(EncodedDict, i)

		dict = append(dict, BEntry{
			Key:   string(key),
			Value: value,
		})

	}

	panic("Dictionary not properly terminated.")
}

func ParseValue(EncodedData []byte, i int) (int, interface{}) {

	switch EncodedData[i] {

	case 'i':
		return ParseInt(EncodedData, i)

	case 'l':
		return ParseList(EncodedData, i)

	case 'd':
		return ParseDict(EncodedData, i)

	default:
		if IsNumeric(EncodedData[i]) {
			return ParseString(EncodedData, i)
		}
	}

	panic("invalid bencode")

}

func spaces(n int) string {
	return fmt.Sprintf("%*s", n, "")
}

func PrettyPrint(v interface{}, indent int) {

	prefix := spaces(indent)
	switch val := v.(type) {

	case int:
		fmt.Printf("%s%d\n", prefix, val)

	case []byte:
		isPrintable := true
		for _, b := range val {
			if b < 32 || b > 126 {
				isPrintable = false
				break
			}
		}

		if isPrintable {
			fmt.Printf("%s%s\n", prefix, string(val))
		} else {
			fmt.Printf("%s<%d bytes binary data>\n", prefix, len(val))
		}

	case string:
		fmt.Printf("%s%s\n", prefix, val)

	case []interface{}:

		fmt.Printf("%s[\n", prefix)
		for _, item := range val {
			PrettyPrint(item, indent+2)
		}
		fmt.Printf("%s]\n", prefix)

	case BDict:
		fmt.Printf("%s{\n", prefix)
		for _, entry := range val {
			fmt.Printf("%s  %s: ", prefix, entry.Key)

			// Inline simple values
			switch entry.Value.(type) {
			case int, string, []byte:
				PrettyPrint(entry.Value, 0)
			default:
				fmt.Println()
				PrettyPrint(entry.Value, indent+4)
			}
		}
		fmt.Printf("%s}\n", prefix)

	default:
		fmt.Printf("%s<unknown>\n", prefix)

	}

}

func test() {
	tests := []struct {
		name  string
		input string
	}{
		{"zero int", "i0e"},
		{"negative int", "i-7e"},
		{"empty string", "0:"},
		{"basic string", "4:spam"},
		{"empty list", "le"},
		{"int list", "li1ei2ei3ee"},
		{"nested list", "ll4:abcdeli99eee"},
		{"empty dict", "de"},
		{"simple dict", "d3:agei25e4:name5:Alicee"},
		{"nested dict", "d4:userd4:name5:Alice3:agei25eee"},
		{"real torrent", "d8:announce35:http://tracker.example.com/announce4:infod4:name8:test.iso6:lengthi1024e12:piece lengthi512e6:pieces21:aaaaabbbbbcccccdddddeee"},
	}

	for _, tt := range tests {
		fmt.Printf("\n=== %s ===\n", tt.name)
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("PANIC: %v\n", r)
				}
			}()
			_, v := ParseValue([]byte(tt.input), 0)
			//fmt.Printf("OK: %#v\n", v)
			PrettyPrint(v, 0)
		}()
	}
}

// Get values from keys
func Get(dict BDict, key string) interface{} {

	for _, entry := range dict {
		if entry.Key == key {
			return entry.Value
		}
	}
	return nil
}

// Build InfoDict
func BuildInfo(info BDict) InfoDict {
	var i InfoDict
	i.Name = string(Get(info, "name").([]byte)) //we know that name is just a slice of bytes
	i.Length = int(Get(info, "length").(int))
	i.PieceLength = Get(info, "piece length").(int)
	val, ok := Get(info, "pieces").([]byte)
	if !ok {
		panic("pieces missing or wrong type")
	}
	i.Pieces = val
	return i

}

// Build Torrent
func BuildTorrent(root BDict) Torrent {
	var t Torrent
	t.Announce = string(Get(root, "announce").([]byte))
	infoRaw := Get(root, "info").(BDict)
	t.Info = BuildInfo(infoRaw)
	return t
}

func main() {

	data, err := os.ReadFile(`C:\Users\Pranjal\Desktop\Projects\BitTorrentClient\archlinux-2026.04.01-x86_64.iso.torrent`)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	//content := string(data)
	//fmt.Println(content)

	_, v := ParseValue(data, 0)
	//PrettyPrint(v, 0)
	d:=v.(BDict)
	t:=BuildTorrent(d)
	fmt.Println(t)

}
