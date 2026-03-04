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

func ParseList(EncodedList string, i int) int {

	//Check  if this is a list at all
	if EncodedList[i] != 'l' {
		panic("This is not a list.")
	}
	i++

	for i < len(EncodedList) {

		ch := EncodedList[i]

		if IsNumeric(ch) {
			i, _ = ParseString(EncodedList, i)

		} else if ch == 'i' {
			i, _ = ParseInt(EncodedList, i)

		} else if ch == 'e' {
			fmt.Println("End of list reached.")
			return (i + 1)
			//break
		} else if ch == 'l' {
			i = ParseList(EncodedList, i)
		}

	}

	panic("List not properly terminated")

}

func ParseDict(EncodedDict string, i int) int {

	//check if this is even a dictionary
	if EncodedDict[i] != 'd' {

		panic("This is not a dictionary.")

	}
	i++

	//Now we loop over the entire dictionary
	for i < len(EncodedDict) {
		ch := EncodedDict[i]

		if IsNumeric(ch) {
			i, _ = ParseString(EncodedDict, i)

		} else if ch == 'i' {
			i, _ = ParseInt(EncodedDict, i)

		} else if ch == 'e' {
			fmt.Println("End of dict reached.")
			return (i + 1)
			//break
		} else if ch == 'l' {
			i = ParseList(EncodedDict, i)
		} else if ch=='d'{
			i=ParseDict(EncodedDict,i)
		}
	}

	return i
}

func main() {

	//EncodedList := "l4:spaml3:hami2eee"
	EncodedDict := "d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee"

	//d4:spaml1:a1:bee represents the dictionary { "spam" => [ "a", "b" ] }
	//ParseList(EncodedList, 0)
	ParseDict(EncodedDict,0)

}
