package main

import (
	"fmt"
)


type BDict []BEntry

type BEntry struct {
    Key   string
    Value interface{}
}

func IsNumeric(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

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

	// Read exact number of bytes
	if i+bytesToRead > len(EncodedString) {
		panic("invalid format: not enough bytes")
	}

	word := EncodedString[i : i+bytesToRead]
	//fmt.Println("Word:", word)
	//fmt.Println("The function returns to point", (i + bytesToRead))

	return (i + bytesToRead), word

}

// returns next index and number
func ParseInt(EncodedInt []byte, i int) (int, int) {
	

	integer := 0

	sign := 1

	if len(EncodedInt) < 3 {

		panic("EncodedInt is too short.")

	}

	if EncodedInt[i] != 'i' {
		panic("Invalid format:No begining delimiter")
	}
	i++

	//Check for negative

	if EncodedInt[i] == '-' {
		sign = -1
		i++

		if i < len(EncodedInt) && EncodedInt[i] == '0' {
            panic("i-0e is not valid bencode")
        }

	}

	//check for leading zeroes
	//li0ee is allowed
	//li03e is not
	if EncodedInt[i]=='0' &&i+1<len(EncodedInt) &&EncodedInt[i+1]!='e' {
		
		panic("Corrupted data : Leading zeroes");
	}

	for i < len(EncodedInt) && EncodedInt[i] != 'e' && EncodedInt[i] != 'i' {
		if !IsNumeric(EncodedInt[i]) {

			panic("Unexpected value in EncodedInt")
		}
		digit := int(EncodedInt[i] - '0')
		integer = integer*10 + digit
		i++

	}

	if EncodedInt[i] != 'e' {
		panic("e was not found at the end of encoded int.")
	}

	integer = sign * integer

	//fmt.Println("int:", integer)
	//fmt.Println("index:", i) //This is the index where the limiting 'e' is present.
	return (i + 1), integer  // we should return the index after the limiting e(when working with lists)
}

func ParseList(EncodedList []byte, i int) (int, []interface{}) {

	var mySlice []interface{}

	//Check  if this is a list at all
	if EncodedList[i] != 'l' {
		panic("This is not a list.")
	}
	i++

	for i<len(EncodedList)&&EncodedList[i]!= 'e' {

		//ch := EncodedList[i]
		var val interface{}
		i,val=ParseValue(EncodedList,i)
		mySlice = append(mySlice, val)
		

	}
	return i+1,mySlice

	

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


func ParseValue(EncodedData[] byte,i int)(int,interface{}){


	switch EncodedData[i]{

	case 'i':
		return ParseInt(EncodedData,i)
	

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


func main() {
    tests := []struct {
        name  string
        input string
    }{
        {"zero int",       "i0e"},
        {"negative int",   "i-7e"},
        {"empty string",   "0:"},
        {"basic string",   "4:spam"},
        {"empty list",     "le"},
        {"int list",       "li1ei2ei3ee"},
        {"nested list",    "ll4:abcdeli99eee"},
        {"empty dict",     "de"},
        {"simple dict",    "d3:agei25e4:name5:Alicee"},
        {"nested dict",    "d4:userd4:name5:Alice3:agei25eee"},
        {"real torrent",   "d8:announce35:http://tracker.example.com/announce4:infod4:name8:test.iso6:lengthi1024e12:piece lengthi512e6:pieces20:aaaaabbbbbcccccdddddee"},
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
            fmt.Printf("OK: %#v\n", v)
        }()
    }
}