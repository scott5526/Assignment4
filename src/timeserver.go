/*
File: timeserver.go
Author: Robinson Thompson

Description: Runs a simple timeserver to pull up a URL page displaying the current time.  Support was verified for Windows 7 OS.  Support has not been tested for other OS

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package main

import (
Log "./seelog-master"
"encoding/json"
"flag"
"fmt"
"html/template"
"io/ioutil"
"math/rand"
"net/http"
"sync"
"os"
//"os/exec"
"strconv"
"time"
)

//var currUser string
var templatesPath *string
var redirect bool
var portNO *int
var printToFile int
var mutex = &sync.Mutex{}
var portInfoStuff PortInfo
var timeout *int
var authport *int
var authhost *string
var backupTime *int
var cookieMap = make(map[string]http.Cookie)
var maxRequests *int
var currentRequests int
var isRequestMax bool
var avgResponse *int
var standardDeviation *int
var runDelay bool

type PortInfo struct {
	PortNum string
}

type TimeInfo struct {
	Name string
	LocalTime string
	UTCTime string
	PortNum string
}

/*
Greeting Redirect 1

Redirects to greetingHandler with a saved URL "/"
*/

func greetingRedirect1(w http.ResponseWriter, r *http.Request) {
    if isRequestMax {
	mutex.Lock()
	currentRequests += 1
	mutex.Unlock()

	if currentRequests > *maxRequests {	
    	    w.WriteHeader(500)
            w.Write([]byte("Internal Service Error"))
        }
    }

    if r.URL.Path != "/" {
	badHandler(w,r) // check if the URL is valid
	return
    }

    fmt.Println("localhost:" + strconv.Itoa(*portNO) + "/")

    if  printToFile == 1 { //Check if the p2f flag is set
	defer Log.Flush()
    	Log.Info("localhost:" + strconv.Itoa(*portNO) + "/\r\n")
    }

    if isRequestMax {
	mutex.Lock()
	currentRequests -= 1
	mutex.Unlock()
    }
    greetingHandler(w,r)
}

/*
Greeting Redirect 2

Redirects to greetingHandler with a saved URL "/index.html"
*/

func greetingRedirect2(w http.ResponseWriter, r *http.Request) {
   if isRequestMax {
	mutex.Lock()
	currentRequests += 1
	mutex.Unlock()

	if currentRequests > *maxRequests {	
    	    w.WriteHeader(500)
            w.Write([]byte("Internal Service Error"))
        }
    }

    fmt.Println("localhost:" + strconv.Itoa(*portNO) + "/index.html")

    if  printToFile == 1 { //Check if the p2f flag is set
	defer Log.Flush()
    	Log.Info("localhost:" + strconv.Itoa(*portNO) + "/index.html\r\n")
    }

    if isRequestMax {
	mutex.Lock()
	currentRequests -= 1
	mutex.Unlock()
    }
    greetingHandler(w,r)
}

/*
Greeting message

Presents the user with a login message if a cookie is found for them, otherwise redirects to the login page
*/
func greetingHandler(w http.ResponseWriter, r *http.Request) {
   if isRequestMax {
	mutex.Lock()
	currentRequests += 1
	mutex.Unlock()

	if currentRequests > *maxRequests {	
    	    w.WriteHeader(500)
            w.Write([]byte("Internal Service Error"))
        }
    }

    greetingCheck(w, r)

    if redirect == true { //If no matching cookie was found in the cookie map, redirect
	path := (*templatesPath + "loginRedirect.html")
    	newTemplate,err := template.New("redirect").ParseFiles(path) 
    	if err != nil {
		fmt.Println("Error running login redirect template")
        	if printToFile == 1 {
			defer Log.Flush()
    			Log.Error("Error running login redirect template\r\n")
		}
		return
    	}   

   	if isRequestMax {
	    mutex.Lock()
	    currentRequests -= 1
	    mutex.Unlock()
        }
    	newTemplate.ExecuteTemplate(w,"loginRedirectTemplate",portInfoStuff)
    }
}

/*
Login handler.  
Displays a html generated login form for the user to provide a name.  
Creates a cookie for the user name and redirects them to the home page if a valid user name was provided.  
If no valid user name was provided, outputs an error message
*/
func loginHandler(w http.ResponseWriter, r *http.Request) {
   if isRequestMax {	
	mutex.Lock()
	currentRequests += 1
	mutex.Unlock()

	if currentRequests > *maxRequests {	
    	    w.WriteHeader(500)
            w.Write([]byte("Internal Service Error"))
        }
    }

    fmt.Println("localhost:" + strconv.Itoa(*portNO) + "/login")

    if  printToFile == 1 { //Check if the p2f flag is set
	defer Log.Flush()
    	Log.Info("localhost:" + strconv.Itoa(*portNO) + "/login\r\n")
    }
  
    loginCheck(w,r)


    // Unique ID generation below

    //tempUUID,_ := exec.Command("uuidgen").Output()
    // uncomment me (^^^^^^^^^) when testing on linux!!!

    newUUID := strconv.Itoa(rand.Int())
    // comment me (^^^^^^^^^) when testing on linux!!!
    //newUUID := string(tempUUID[:])
    // uncomment me (^^^^^^^^^) when testing on linux!!!

    expDate := time.Now()
    expDate.AddDate(1,0,0)
    cookieSetup(w, r, newUUID, expDate)

    if isRequestMax {
	 mutex.Lock()
	 currentRequests -= 1
	 mutex.Unlock()
    }
}

/*
Logout handler.  

Clears user cookie, displays goodbye message for 10 seconds, then redirects user to login form
*/
func logoutHandler(w http.ResponseWriter, r *http.Request) {
   if isRequestMax {
	mutex.Lock()
	currentRequests += 1
	mutex.Unlock()

	if currentRequests > *maxRequests {	
    	    w.WriteHeader(500)
            w.Write([]byte("Internal Service Error"))
        }
    }

   fmt.Println("localhost:" + strconv.Itoa(*portNO) + "/logout")

    if  printToFile == 1 { //Check if the p2f flag is set
	defer Log.Flush()
    	Log.Info("localhost:" + strconv.Itoa(*portNO) + "/logout\r\n")
    }

   clearMapCookie(r)

    // User wasn't actually logged in, redirect them to login page
    if !redirect {
	path := (*templatesPath + "loginRedirect.html")
    	newTemplate,err := template.New("redirect").ParseFiles(path) 
    	if err != nil {
		fmt.Println("Error running login redirect template")
        	if printToFile == 1 {
			defer Log.Flush()
    			Log.Error("Error running login redirect template\r\n")
		}
		return
    	}   
        if isRequestMax {
	    mutex.Lock()
	    currentRequests -= 1
	    mutex.Unlock()
        }
    	newTemplate.ExecuteTemplate(w,"loginRedirectTemplate",portInfoStuff)
    }

    //Redirect to the login page
    path := *templatesPath + "logoutToLoginRedirect.html"
    newTemplate,err := template.New("redirect").ParseFiles(path)  
    if err != nil {
	fmt.Println("Error running login redirect template")
        if printToFile == 1 {
		defer Log.Flush()
    		Log.Error("Error running login redirect template\r\n")
	}
	return;
    }  

    if isRequestMax {
	mutex.Lock()
	currentRequests -= 1
	mutex.Unlock()
    }
    newTemplate.ExecuteTemplate(w,"loginRedirectTemplate",portInfoStuff)
}


/*
Handler for time requests.  

Outputs the current time in the format:
Hour:Minute:Second PM/AM
*/
func timeHandler(w http.ResponseWriter, r *http.Request) {
    if isRequestMax {
	mutex.Lock()
	currentRequests += 1
	mutex.Unlock()

	if currentRequests > *maxRequests {	
    	    w.WriteHeader(500)
            w.Write([]byte("Internal Service Error"))
        }
    }

    fmt.Println("localhost:" + strconv.Itoa(*portNO) + "/time")

    if  printToFile == 1 { //Check if the p2f flag is set
	defer Log.Flush()
    	Log.Info("localhost:" + strconv.Itoa(*portNO) + "/time\r\n")
    }

    user := getUserName(r)

    currTime := time.Now().Format("03:04:05 PM")
    utcTime := time.Now().UTC()
    utcTime = time.Date(
        time.Now().UTC().Year(),
        time.Now().UTC().Month(),
        time.Now().UTC().Day(),
        time.Now().UTC().Hour(),
        time.Now().UTC().Minute(),
        time.Now().UTC().Second(),
        time.Now().UTC().Nanosecond(),
        time.UTC,
    )

    utcTime.UTC()
    //utcTime.Format("03:04:05 07")

    //setup time struct for time.html
    currTimeInfo := TimeInfo {
    	Name: user,
	LocalTime: currTime,
	UTCTime: utcTime.Format("03:04:05"),
	PortNum: strconv.Itoa(*portNO),
    }

    path := *templatesPath + "time.html"
    newTemplate,err := template.New("timeoutput").ParseFiles(path)  
    if err != nil {
	fmt.Println("Error running time template")
        if printToFile == 1 {
		defer Log.Flush()
    		Log.Error("Error running time template\r\n")
	}
	return;
    } 

    if runDelay {
	time.Sleep(time.Duration(generateDelay())*(time.Second/1000))
    }
    
    newTemplate.ExecuteTemplate(w,"timeTemplate",currTimeInfo)
    if isRequestMax {
	mutex.Lock()
	currentRequests -= 1
	mutex.Unlock()
    }
}

/*
Generates a random relay based on the standard deviation & average response time flags
*/
func generateDelay() int {
    deviation := (rand.Int() % 1000) + 1
    change := ((rand.Int() % 100) + 1) / 100
    change = change * *standardDeviation
    addorMinus := rand.Int() % 2
    delay := 0

    if deviation < 682 { //within 1st standard deviation of avg
	if addorMinus == 0 { // add
	    delay = *avgResponse + change
        } else { // minus
	    delay = *avgResponse - change

	    if delay < 0 {
		delay = 0
	    }
        }
    } else if deviation < 958 { //within 2nd standard deviation of avg
	if addorMinus == 0 { // add
	    delay = *avgResponse + *standardDeviation + change
        } else { // minus
	    delay = *avgResponse - *standardDeviation - change

	    if delay < 0 {
		delay = 0
	    }
        }
    } else { //within 3rd standard deviation of avg
	if addorMinus == 0 { // add
	    delay = *avgResponse + (*standardDeviation * 2) + change
        } else { // minus
	    delay = *avgResponse - (*standardDeviation * 2) - change

	    if delay < 0 {
		delay = 0
	    }
        }
    }

    return delay
}

/*
Menu handler.  

Displays menu consisting of Home, Time, Logout, and About us
*/
func menuHandler(w http.ResponseWriter, r *http.Request) {
   if isRequestMax {
	mutex.Lock()
	currentRequests += 1
	mutex.Unlock()

	if currentRequests > *maxRequests {	
    	    w.WriteHeader(500)
            w.Write([]byte("Internal Service Error"))
        }
    }

   fmt.Println("localhost:" + strconv.Itoa(*portNO) + "/menu")

   if  printToFile == 1 { //Check if p2f flag is set
	defer Log.Flush()
    	Log.Info("localhost:" + strconv.Itoa(*portNO) + "/menu\r\n")
   }

    //Open the menu page
    path := *templatesPath + "menu.html"
    newTemplate,err := template.New("open").ParseFiles(path)  
    if err != nil {
	fmt.Println("Error running menu HTML template")
        if printToFile == 1 {
		defer Log.Flush()
    		Log.Error("Error running menu redirect template\r\n")
	}
	return;
    }  
    newTemplate.ExecuteTemplate(w,"menuTemplate",portInfoStuff)

    if isRequestMax {
	mutex.Lock()
	currentRequests -= 1
	mutex.Unlock()
    }
}

/*
Handler for invalid requests.  Outputs a 404 error message and a cheeky message
*/
func badHandler(w http.ResponseWriter, r *http.Request) {
   if isRequestMax {
	mutex.Lock()
	currentRequests += 1
	mutex.Unlock()

	if currentRequests > *maxRequests {	
    	    w.WriteHeader(500)
            w.Write([]byte("Internal Service Error"))
        }
    }

    if r.URL.Path == "/index.html" {
	return
    } else if r.URL.Path == "/login" {
	return
    } else if r.URL.Path == "/logout" {
	return
    } else if r.URL.Path == "/time" {
	return
    } else if r.URL.Path == "/menu" {
	return
    }

    http.NotFound(w, r)
    w.Write([]byte("These are not the URLs you're looking for."))

    if isRequestMax {
	mutex.Lock()
	currentRequests -= 1
	mutex.Unlock()
    }
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
    fmt.Println("Starting new server")

    //Version output & port selection
    version := flag.Bool("V", false, "Version 4.2") //Create a bool flag for version  
    						    //and default to no false

    portNO = flag.Int("port", 8080, "")	    	    //Create a int flag for port selection
					            //and default to port 8080

    p2f := flag.Bool("p2f", false, "") //flag to output to file

    templatesPath = flag.String("templates", "Templates/", "")

    logPath := flag.String("log", "seelog.xml", "")

    timeout = flag.Int("authtimeout-ms", 2000, "")

    authport = flag.Int("authport", 9090, "")

    authhost = flag.String("authhost", "localhost:", "")
     
    loadDumpFile := flag.Bool("dumpfile", false, "")

    backupTime = flag.Int("checkpoint-interval", 0, "")

    maxRequests = flag.Int("max-inflight", 0, "")

    avgResponse = flag.Int("avg-response-ms", -1, "")

    standardDeviation = flag.Int("deviation-ms", -1, "")

    //Setup the seelog logger (cudos to http://sillycat.iteye.com/blog/2070140, https://github.com/cihub/seelog/blob/master/doc.go#L57)
    logger,loggerErr := Log.LoggerFromConfigAsFile("../etc/" + *logPath)
    if loggerErr != nil {
    	fmt.Println("Error creating logger from .xml log configuration file")
    } else {
	Log.ReplaceLogger(logger)
    }
    

    printToFile = 0 // set to false
    flag.Parse()

    if *version == true {		//If version outputting selected, output version and 
        fmt.Println("Version 4.2")	//terminate program with 0 error code
        os.Exit(0)
    }

    if *p2f == true {
	printToFile = 1 // set to true
    }
	
    portInfoStuff = PortInfo{
	PortNum: strconv.Itoa(*portNO),
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

    //Indicate whether or not there is a maximum # of concurrent requests
    if *maxRequests > 0 {
	isRequestMax = true
    } else {
	isRequestMax = false
    }

    currentRequests = 0

    //Need avgResponse AND standardDeviation or don't use either
    runDelay = true
    if *avgResponse < 0 || *standardDeviation < 0 {
	runDelay = false
    } else if *avgResponse < 0 {
	fmt.Println("avg-response-ms flag used, but not deviation-ms.  avg-response-ms being ignored")
	if printToFile == 1 {
	    defer Log.Flush()
	    Log.Warn("avg-response-ms flag used, but not deviation-ms.  avg-response-ms being ignored")
	}
    } else if *standardDeviation < 0 {
	fmt.Println("deviation-ms flag used, but not avg-response-ms.  deviation-ms being ignored")
	if printToFile == 1 {
	    defer Log.Flush()
	    Log.Warn("deviation-ms flag used, but not avg-response-ms.  deviation-ms being ignored")
	}	
    }

    // URL handling
    http.HandleFunc("/", greetingRedirect1)
    http.HandleFunc("/index.html", greetingRedirect2)
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/logout", logoutHandler)
    http.HandleFunc("/time", timeHandler)
    http.HandleFunc("/menu", menuHandler)
    
    //Check host:(specified port #) for incomming connections
    error := http.ListenAndServe("localhost:" + strconv.Itoa(*portNO), nil)

    if error != nil {				// If the specified port is already in use, 
	fmt.Println("Port already in use")	// output a error message and exit with a 
        if printToFile == 1 {
		defer Log.Flush()
    		Log.Error("Port already in use\r\n")
        }
	os.Exit(1)				// non-zero error code
    }
}
