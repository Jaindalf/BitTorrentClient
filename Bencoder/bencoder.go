package main

import (
	"fmt"
	"time"
)

func IsNumeric(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func ParseString(EncodedString string) (int,string) {

	bytesToRead := 0
	i := 0

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

	return (i+bytesToRead),word

}
//returns next index and number
func ParseInt(EncodedInt string) (int,int) {

	integer := 0
	i := 0
	sign := 1 //0 for pos and 1  for negative

	if len(EncodedInt)<3 {

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

	if EncodedInt[i] == '0' &&len(EncodedInt)>3{
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
	return (i),integer  // we should return the index  of  the next element 
}

func ParseList(EncodedList string) {
	i:=0

	//Check  if this is a list at all
	if EncodedList[i]!='l'{
		panic("This is not a list.")
	}
	i++

	for i<len(EncodedList){
		fmt.Println("The value of i is:",i)
		time.Sleep(3 * time.Second) 


		ch:=EncodedList[i]

		if IsNumeric(ch) {
			i,_=ParseString(EncodedList[i:])
			fmt.Println("The value of i(below parsestring) is:",i)
			
		} else if ch=='i'{
			i,_=ParseInt(EncodedList[i:])
			fmt.Println("The value of i(below parse int) is:",i)
		}

		//fmt.Println("The value of i is:",i)


	}
	


}

func main() {
	//EncodedString := "5:sphami3e"
	//EncodedInt := "i2e"
	
	//ParseString(EncodedString)
	//ParseInt(EncodedInt)
	//EncodedList:="li5e3:tyre" //equivalent to ["spam",34]
	//ParseList(EncodedList)
	ei:="li2e3"
	index,integer:=(ParseInt(ei[1:]))

	fmt.Println("INDEX RETURNED",index)
	fmt.Println("VALUE AT RETURNED INDEX:",string(ei[index]))
	fmt.Println("INTEGER RETURNED:",integer)
}
