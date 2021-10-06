package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)
// Printas ut i http://localhost:8000/

const URLprefix = "/objects/"

// Allows you to either POST (put) data and to GET (get) data from json HTTP requests.
func (network *Network) HTTPhandler(w http.ResponseWriter, r *http.Request){
	switch r.Method {
	case "POST":
		body,error := ioutil.ReadAll(r.Body) // Read Request
		defer r.Body.Close() // Always CLOSE.

		cleanBody := removeQuotationMarks(string(body)) // Turns " "test" " into "  test   "

		// Check for errors or if body is empty.
		if error != nil || string(cleanBody) == "" {
			http.Error(w, "ERROR", http.StatusBadRequest)
			fmt.Println("Error when POST")
		}  else{
			// Same as in Cli.go Store
			hashedFileString := NewKademliaIDFromData(cleanBody)
			network.Store([]byte(body),hashedFileString)

			// Done as per the lab instructions.
			hashSuffix := hashedFileString.String()
			w.Header().Set("Location", URLprefix+hashSuffix)
			w.WriteHeader(http.StatusCreated)	// Status 201 as detailed.
			w.Write(body)
		}
	case "GET":
		// Checks if there is something after the prefix.  /objects/XXXXXXXXXXXXXX
		URLcomponents := strings.Split(r.URL.Path, "/")	// [ "", "objects", "hash" ]
		hashValue := URLcomponents[2]

		// Check if there is a hashvalue of correct size.
		if(len(hashValue) != 40){
			http.Error(w, "ERROR", http.StatusLengthRequired)
			fmt.Println("Error when GET ", hashValue, " is not of correct length. (40)")
		}else{
				// Same as in Cli.go Get
				hash := NewKademliaID(hashValue)
				data, nodes := network.DataLookup(hash)
				if data != nil {
					// If data is not nil, send OK status and write.
					w.WriteHeader(http.StatusOK)
					w.Write(data)
				} else if len(nodes) > 0{
					http.Error(w, "ERROR", http.StatusNotFound)
					fmt.Println("Error when GET - DataLookUP (Length)")
				} else {
					http.Error(w, "ERROR", http.StatusNoContent)
					fmt.Println("Error when GET - DataLookUP")
				}
		}
	default:
		http.Error(w, "Wrong. Use POST or GET", http.StatusMethodNotAllowed)
	}
}

// Could need some work.

// Enables listening to HTTP
func (network *Network) HTTPlisten() {
	// https://github.com/gorilla/mux
	filePath := "/objects/" // Specified in lab.
	r := mux.NewRouter()
	r.HandleFunc(filePath, network.HTTPhandler)
	log.Fatal(http.ListenAndServe(":8000", r))
}
// Remove first and last char of string (Quotation Marks)
func removeQuotationMarks(str string) string {
	stringStart := 1
	stringEnd := len(str)-1
	return str[stringStart : stringEnd]
}
