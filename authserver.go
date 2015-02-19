/*
File: authserver.go
Author: Robinson Thompson

Description: Authentication server for timeserver to get/set user cookies

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package auth

import (
Log "./seelog-master"
"flag"
"fmt"
"net/http"
"strconv"
)
var authport int
var hostname string

/*
Handler for cookie "get" requests
Attmpts to lookup user name based on provided cookie UUID.  
Returns user name or empty with response code 200 on valid request.  
Returns empty string with response code 400 on malformed request
*/
func getRedirectHandler (w http.ResponseWriter, r *http.Request) string {
    cookieName := ""
    cookieUUID := r.FormValue("cookie")
    if cookieUUID == "" { 
	w.WriteHeader(400) // set response code to 400, malformed request
	return cookieName
    }
     
    //Attempt to retrieve user name from cookie map based on UUID

    w.WriteHeader(200) // set response code to 200, request processed
    return cookieName    
}

/*
Handler for cookie "set" requests
Attempts to add a new provided cookie to internal cookie map
Returns response code 200 on processed request
Returns response code 400 on malformed request
*/
func setRedirectHandler (w http.ResponseWriter, r *http.Request) {
    cookie := r.FormValue("cookie")
    cookieName := r.FormValue("name")
    if cookie == nil {
	w.WriteHeader(400) // set response code to 400, malformed request
	return
    }
    if cookieName == "" {
	w.WriteHeader(400) // set response code to 400, malformed request
	return
    }

    // attempt to add cookie to internal cookie map


    w.WriteHeader(200) // set response code to 200, request processed
}


/*

*/
func errorHandler (w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(404) // Set response code to 404 not found
}

/*
Main
*/
func main() {
    authport = flag.Int("authport", 8181, "")
    hostname = flag.String("hostname", "localhost:", "")

    http.HandleFunc("/get?cookie", getRedirectHandler)
    http.HandleFunc("/set?cookie", setRedirectHandler)
    http.HandleFunc("/", errorHandler)

    error := http.ListenAndServe(hostname + strconv.Itoa(authport), nil)
    if error != nil {				// If the specified port is already in use, 
	fmt.Println("Port already in use")	// output a error message and exit with a 
        if printToFile == 1 {
		defer Log.Flush()
    		Log.Error("Port already in use\r\n")
        }
	os.Exit(1)				// non-zero error code
    }
}
