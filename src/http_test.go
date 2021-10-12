package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Tests for POST, GET and INVALID inputs through HTTP Requests.
func TestHTTPhandler(t *testing.T) {
	// Setup
	addrs,_ := net.InterfaceAddrs()
	var testIP net.IP
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				testIP = ipnet.IP
			}
		}
	}
	prefix := "/objects/"
	net:= NewNetwork(&testIP,NewMessageService(false,nil))

	// POST valid
	httpRecorder1 := httptest.NewRecorder()
	input1 := "test"
	jsonInput1, _ := json.Marshal(input1)
	request1 := httptest.NewRequest("POST", ("/objects"), bytes.NewBuffer(jsonInput1))
	request1.Close =true

	net.HTTPhandler(httpRecorder1, request1)
	status1 := httpRecorder1.Code
	expectedStatus1 := http.StatusCreated

	if(status1 != expectedStatus1){
		t.Errorf("WRONG STATUS CODE: GOT %v EXPECTED %v", status1, expectedStatus1)
	}else{
		fmt.Println("HTTP - POST Valid Input = Passed")
	}

	// Get Valid
	httpRecorder2 := httptest.NewRecorder()
	input2 := "5006d6f8302000e8b87fef5c50c071d6d97b4e88"
	request2 := httptest.NewRequest("GET", (prefix+input2),nil)
	request2.Close =true

	net.HTTPhandler(httpRecorder2, request2)
	status2 := httpRecorder2.Code
	expectedStatus2 := http.StatusOK

	if(status2 != expectedStatus2){
		t.Errorf("WRONG STATUS CODE: GOT %v EXPECTED %v", status2, expectedStatus2)
	}else{
		fmt.Println("HTTP - GET Valid Input = Passed")
	}

	// Invalid Method Name
	httpRecorder3 := httptest.NewRecorder()
	input3 := ""
	request3 := httptest.NewRequest("KADEMLIA", (prefix+input3), nil)
	request3.Close =true

	net.HTTPhandler(httpRecorder3, request3)
	status3 := httpRecorder3.Code
	expectedStatus3 := http.StatusMethodNotAllowed

	if(status3 != expectedStatus3){
		t.Errorf("WRONG STATUS CODE: GOT %v EXPECTED %v", status3, expectedStatus3)
	}else{
		fmt.Println("HTTP - Invalid Method Name = Passed")
	}
	// POST Invalid (Empty String)
	httpRecorder4 := httptest.NewRecorder()
	input4 := ""
	jsonInput4, _ := json.Marshal(input4)
	request4 := httptest.NewRequest("POST", (prefix+input4), bytes.NewBuffer(jsonInput4))
	request4.Close =true

	net.HTTPhandler(httpRecorder4, request4)
	status4 := httpRecorder4.Code
	expectedStatus4 := http.StatusBadRequest

	if(status4 != expectedStatus4){
		t.Errorf("WRONG STATUS CODE: GOT %v EXPECTED %v", status4, expectedStatus4)
	}else{
		fmt.Println("HTTP - POST Invalid Input (Empty String) = Passed")
	}

	// Get Invaid (Bad length of Hash)
	httpRecorder5 := httptest.NewRecorder()
	input5 := "thisisnothash!"
	request5 := httptest.NewRequest("GET", (prefix+input5), nil)
	request5.Close =true

	net.HTTPhandler(httpRecorder5, request5)
	status5 := httpRecorder5.Code
	expectedStatus5 := http.StatusLengthRequired

	if(status5 != expectedStatus5){
		t.Errorf("WRONG STATUS CODE: GOT %v EXPECTED %v", status5, expectedStatus5)
	}else{
		fmt.Println("HTTP - GET Invalid Input (Len) = Passed")
	}
	// Get Invaid (Non-Existing Hash)
	httpRecorder6 := httptest.NewRecorder()
	input6 := "a94a8fe5ccb19ba61c4c0873d391e98798200000"
	request6 := httptest.NewRequest("GET", (prefix+input6), nil)
	request6.Close =true

	net.HTTPhandler(httpRecorder6, request6)
	status6 := httpRecorder6.Code
	expectedStatus6 := http.StatusNoContent

	if(status6 != expectedStatus6){
		t.Errorf("WRONG STATUS CODE: GOT %v EXPECTED %v", status6, expectedStatus6)
	}else{
		fmt.Println("HTTP - GET Valid Input = Passed")
	}
}