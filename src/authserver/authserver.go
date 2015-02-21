/*
File: authserver.go
Author: Robinson Thompson

Description: Authentication server for timeserver to get/set user cookies

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package main

import (
Log "../seelog-master"
"encoding/json"
"io/ioutil"
"flag"
"fmt"
"net/http"
"os"
"strconv"
"sync"
"time"
)
var authport *int
var hostname *string
var backupTime *int
var printToFile int
var mutex = &sync.Mutex{}
var cookieMap = make(map[string]http.Cookie)

/*
Handler for cookie "get" requests
Attmpts to lookup user name based on provided cookie UUID.  
Returns user name or empty with response code 200 on valid request.  
Returns empty string with response code 400 on malformed request
*/
func getRedirectHandler (w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    cookieName := ""
    cookieUUID := r.URL.Path.Get("cookie")
    if cookieUUID == "" { 
	w.WriteHeader(400) // set response code to 400, malformed request
    }
     
    //Attempt to retrieve user name from cookie map based on UUID
    foundCookie := false

    mutex.Lock()
    for _, currCookie := range cookieMap {  //Run through the range of applicable cookies on the user's browser
    	if (currCookie.Name != "") {
	    currCookieVal := currCookie.Name

            if (currCookieVal == cookieUUID) {
		foundCookie = true
		cookieName = currCookie.Value
	    }
    	}
     }
     mutex.Unlock()

     if !foundCookie {
	 w.WriteHeader(400) // set response code to 400, malformed request
     }

     w.WriteHeader(200) // set response code to 200, request processed
     // redirect to timeserver URL localhost:8080/get?name=cookieName 
}

/*
Handler for cookie "set" requests
Attempts to add a new provided cookie to internal cookie map
Returns response code 200 on processed request
Returns response code 400 on malformed request
*/
func setRedirectHandler (w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    cookie := r.URL.Path.Get("cookie")
    cookieName := r.URL.Path.Get("name")
    if cookie == "" || cookieName == "" {
	w.WriteHeader(400) // set response code to 400, request malformed
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
Updates dumpfile.txt 
*/
func Updatedumpfile() {
    stallDuration := time.Duration(*backupTime)*time.Second
    for {
	time.Sleep(stallDuration)
        mutex.Lock()
        encodedMap,_ := json.Marshal(cookieMap)
        mutex.Unlock()

        oldDump,err := ioutil.ReadFile("dumpfile.txt")
        if err != nil { //Assume that dumpfile.txt hasn't been made yet
            ioutil.WriteFile("dumpfile.txt", encodedMap, 0644)
	    readCopy,err2 := ioutil.ReadFile("dumpfile.txt")
	    if err2 != nil {
	       fmt.Println("Error reading dumpfile")
                if printToFile == 1 {
		    defer Log.Flush()
		    Log.Error("Error reading dumpfile")
		    return
	        }
            }
	    mutex.Lock()
	    err3 := json.Unmarshal(readCopy, &cookieMap)
	    mutex.Unlock()
	    if err3 != nil {
	        fmt.Println("Error unmarshaling")
	        if printToFile == 1 {
	    	    defer Log.Flush()
		    Log.Error("Error unmarshaling")
	        }
            }
	    return
        }

        ioutil.WriteFile("dumpfile.bak", oldDump, 0644)
        os.Remove("dumpfile.txt")
        ioutil.WriteFile("dumpfile.txt", encodedMap, 0644)
        readCopy,err5 := ioutil.ReadFile("dumpfile.txt")
        if err5 != nil {
	    fmt.Println("Error reading dumpfile")
            if printToFile == 1 {
	        defer Log.Flush()
	        Log.Error("Error reading dumpfile")
	        return
	    }
        }

        mutex.Lock()
        err6 := json.Unmarshal(readCopy, &cookieMap)
        mutex.Unlock()
        if err6 != nil {
	    fmt.Println("Error unmarshaling")
	    if printToFile == 1 {
                defer Log.Flush()
	        Log.Error("Error unmarshaling")
	    }
	    return
        }
        os.Remove("dumpfile.bak")
    }
}

/*
Main
*/
func main() {
    p2f := flag.Bool("p2f", false, "") //flag to output to file

    logPath := flag.String("log", "seelog.xml", "")

    authport = flag.Int("authport", 8181, "")

    hostname = flag.String("hostname", "localhost:", "")

    loadDumpFile := flag.Bool("dumpfile", false, "")

    backupTime = flag.Int("checkpoint-interval", 0, "")

    printToFile = 0 // set to false
    if *p2f == true {
	printToFile = 1 // set to true
    }


    if *loadDumpFile == true {
	dump,err := ioutil.ReadFile("dumpfile.txt")
	if err != nil {
	    fmt.Println("Error reading dumpfile")
            if printToFile == 1 {
		defer Log.Flush()
		Log.Error("Error reading dumpfile")
	    }
        } else {
	    mutex.Lock()
	    err2 := json.Unmarshal(dump, &cookieMap)
	    mutex.Unlock()
	    if err2 != nil {
	        fmt.Println("Error unmarshaling")
	        if printToFile == 1 {
		    defer Log.Flush()
		    Log.Error("Error unmarshaling")
	        }
            }
        }
    }

    if *backupTime > 0 {
	go Updatedumpfile()
    }

    //Setup the seelog logger (cudos to http://sillycat.iteye.com/blog/2070140, https://github.com/cihub/seelog/blob/master/doc.go#L57)
    logger,loggerErr := Log.LoggerFromConfigAsFile("../../etc/" + *logPath)
    if loggerErr != nil {
    	fmt.Println("Error creating logger from .xml log configuration file")
    } else {
	Log.ReplaceLogger(logger)
    }

    http.HandleFunc("/get?cookie", getRedirectHandler)
    http.HandleFunc("/set?cookie", setRedirectHandler)
    http.HandleFunc("/", errorHandler)

    error := http.ListenAndServe(*hostname + strconv.Itoa(*authport), nil)
    if error != nil {				// If the specified port is already in use, 
	fmt.Println("Port already in use")	// output a error message and exit with a 
        if printToFile == 1 {
		defer Log.Flush()
    		Log.Error("Port already in use\r\n")
        }
	os.Exit(1)				// non-zero error code
    }
}
