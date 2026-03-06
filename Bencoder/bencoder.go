package main

import (
	"fmt"
)

func IsNumeric(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func ParseString(EncodedString string, i int) (int, string) {

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

	fmt.Println("Length:", bytesToRead)

	// Read exact number of bytes
	if i+bytesToRead > len(EncodedString) {
		panic("invalid format: not enough bytes")
	}

	word := EncodedString[i : i+bytesToRead]
	fmt.Println("Word:", word)
	fmt.Println(string(EncodedString[i+bytesToRead]))
	fmt.Println("The function returns to point", (i + bytesToRead))

	return (i + bytesToRead), word

}

// returns next index and number
func ParseInt(EncodedInt string, i int) (int, int) {

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

	}

	//check for leading zeroes

	if EncodedInt[i] == '0' && len(EncodedInt) > 3 {
		fmt.Println("Leading  zeroes not  allowed.")
		panic("Corrupted  data")
	}

	for i < len(EncodedInt) && EncodedInt[i] != 'e' && EncodedInt[i] != 'i' {
		if !IsNumeric(EncodedInt[i]) {

			panic("Unexppected value in EncodedInt")
		}
		digit := int(EncodedInt[i] - '0')
		integer = integer*10 + digit
		i++

	}

	if EncodedInt[i] != 'e' {
		panic("e was not found at the end of encoded int.")
	}

	integer = sign * integer

	fmt.Println("int:", integer)
	fmt.Println("index:", i) //This is the index where the limiting 'e' is present.
	return (i + 1), integer  // we should return the index after the limiting e(when working with lists)
}

func ParseList(EncodedList string, i int) (int, []interface{}) {

	var mySlice []interface{}

	//Check  if this is a list at all
	if EncodedList[i] != 'l' {
		panic("This is not a list.")
	}
	i++

	for i < len(EncodedList) {

		ch := EncodedList[i]

		if IsNumeric(ch) {
			var word string
			i, word = ParseString(EncodedList, i)
			mySlice = append(mySlice, word)

		} else if ch == 'i' {
			var integer int
			i, integer = ParseInt(EncodedList, i)
			mySlice = append(mySlice, integer)

		} else if ch == 'e' {
			fmt.Println("End of list reached.")
			return (i + 1), mySlice
			//break
		} else if ch == 'l' {
			var innerSlice []interface{}
			i, innerSlice = ParseList(EncodedList, i)
			mySlice = append(mySlice, innerSlice)
		}

	}

	panic("List not properly terminated")

}

func ParseDict(EncodedDict string, i int) (int, map[string]interface{},[]interface{}) {

	//Dictionaries are encoded as follows: d<bencoded string><bencoded element>e
	// the key can only be a string

	//check if this is even a dictionary
	if EncodedDict[i] != 'd' {

		panic("This is not a dictionary.")

	}
	i++

	dict := make(map[string]interface{})

	var keys []interface{}

	//Now we loop over the entire dictionary
	for i < len(EncodedDict) {

		// End of dictionary
		if EncodedDict[i] == 'e' {
			fmt.Println("End of dict reached.")
			return i + 1, dict,keys
		}

		// Parse key (always a string)
		var key string
		i, key = ParseString(EncodedDict, i)
		fmt.Println("KEY:", key)
		keys = append(keys, key)

		var value interface{}
		ch := EncodedDict[i]

		if IsNumeric(ch) {
			i, value = ParseString(EncodedDict, i)

		} else if ch == 'i' {
			i, value = ParseInt(EncodedDict, i)

		} else if ch == 'e' {
			fmt.Println("End of dict reached.")
			return (i + 1), dict,keys
			//break
		} else if ch == 'l' {
			i, value = ParseList(EncodedDict, i) ///
		} else if ch == 'd' {
			var innerkeys []interface{}
			fmt.Println("parsing dict.")
			i, value,innerkeys = ParseDict(EncodedDict, i)
			keys=append(keys, innerkeys)
			
		}
		// Store key-value pair
		dict[key] = value
		

	}

	panic("Dictionary not properly terminated.")
}



func main() {

	//EncodedDict := "d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:home3:keyi73e4:prani339999ee"
	//EncodedDict := "d4:userd4:name5:Alice3:agei25ee6:status6:onlinee"
	//Go doesn't have ordered maps  ༽◺_◿༼ ༽◺_◿༼ ༽◺_◿༼
		str := "d8:announce41:http://bttracker.debian.org:6969/announce7:comment35:\"Debian CD from cdimage.debian.org\"13:creation datei1391870037e9:httpseedsl85:http://cdimage.debian.org/cdimage/release/7.4.0/iso-cd/debian-7.4.0-amd64-netinst.iso85:http://cdimage.debian.org/cdimage/archive/7.4.0/iso-cd/debian-7.4.0-amd64-netinst.isoe4:infod6:lengthi232783872e4:name30:debian-7.4.0-amd64-netinst.iso12:piece lengthi262144e6:pieces0:ee"

	_, _,keys := ParseDict(str, 0)
	fmt.Println(keys)

	


}
