package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)


func TestSha1Hash(t *testing.T) {
	fmt.Println("This import is not unused >:(")
}

func TestNewKademliaID(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	input := ""
	for i := 0; i < ID_LEN; i++ {
		r := rand.Intn(100 - 10) + 10
		input += strconv.Itoa(r)
	}
	id := NewKademliaID(input)
	if id.String() != input {
		t.Errorf("NewKademliaID got %s, expected %s", id.String(), input)
	}
}

func TestNewKademliaIDFromData(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var input string
	for i := 0; i < ID_LEN; i++ {
		r := rand.Intn(11)
		input = strconv.Itoa(r)
	}
	id := NewKademliaIDFromData(input)
	res, _ := hex.DecodeString(id.String())

	for i := 0; i < ID_LEN; i++ {
		if res[i] != sha1Hash(input)[i] {
			t.Errorf("NewKademliaIDFromData error at byte array index %d", i)
		}
	}
}

func TestNewKademliaIDFromIP(t *testing.T) {
	fmt.Println("This import is not unused >:(")

}

func TestKademliaID_Less(t *testing.T) {
	fmt.Println("This import is not unused >:(")

}

func TestKademliaID_Equals(t *testing.T) {
	fmt.Println("This import is not unused >:(")

}

func TestKademliaID_CalcDistance(t *testing.T) {
	fmt.Println("This import is not unused >:(")

}

func TestKademliaID_String(t *testing.T) {
	fmt.Println("This import is not unused >:(")

}
