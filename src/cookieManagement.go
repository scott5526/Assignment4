/*
File: cookieManagement.go
Author: Robinson Thompson

Description:  Manages cookies for timeserver.go

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package main

import (
Log "./seelog-master"
"fmt"
"html/template"
"net/http"
"strconv"
"time"
)

//Adds a cookie to the cookie map
func mapSetCookie (newCookie http.Cookie, newUUID string) {
	mutex.Lock()
	cookieMap[newUUID] = newCookie
	mutex.Unlock()
}

//Attempt to find a cookie on the user's browser and greet them using their name stored on it
func greetingCheck (w http.ResponseWriter, r *http.Request) {
    redirect = true
    for _, currCookie := range r.Cookies() { // check all potential cookies stored by the user for a matching cookie
    	if (currCookie.Name != "") {
	    currCookieVal := currCookie.Value
	    mutex.Lock()
	    mapCookie := cookieMap[currCookieVal]
	    mutex.Unlock()
            if (mapCookie.Value != "") {
    		fmt.Fprintf(w, "Greetings, " + mapCookie.Value)
		redirect = false
	    }
	}
    }
}

//Ensuring the user does not already have a browser cookie matching a cookie in the local cookie map, if they do
//redirect the user to the greetings page
func loginCheck (w http.ResponseWriter, r *http.Request) {
    for _, currCookie := range r.Cookies() {  //Run through the range of applicable cookies on the user's browser
    	if (currCookie.Name != "") {
	currCookieVal := currCookie.Value
	mutex.Lock()
	mapCookie := cookieMap[currCookieVal]  //Find the corresponding cookie in the local cookie map
	mutex.Unlock()
        	if (mapCookie.Value != "") {
			path := *templatesPath + "greetingRedirect.html"
    			newTemplate,err := template.New("redirect").ParseFiles(path)  
    			if err != nil {
				fmt.Println("Error running greeting redirect template")
				return
    			}  
    			newTemplate.ExecuteTemplate(w,"greetingRedirectTemplate",portInfoStuff)
		}
    	}
     }
}

//Clears a user's cookie from the cookie map and clear (1) copy of the cookie from the user's browser (cannot delete
//copies the user may have created themselves and stored elsewhere)
func clearMapCookie (r *http.Request) {
   redirect = false // set to true if user cookie is found (they are actually logged in)
   for _, currCookie := range r.Cookies() {  //Run through the range of applicable cookies on the user's browser
    	if (currCookie.Name != "") {
	currCookieVal := currCookie.Value
	mutex.Lock()
	mapCookie := cookieMap[currCookieVal]  //Find the corresponding cookie in the local cookie map
	mutex.Unlock()
        	if (mapCookie.Value != "") {
			redirect = true // user was actually logged in
			mutex.Lock()
    			delete(cookieMap, currCookieVal) //Delete the cleared cookie from the local cookie map
			mutex.Unlock()
			currCookie.MaxAge = -1 //Set the user's cookie's MaxAge to an invalid number to expire it
		}
    	}
    }
}

//Creates a cookie for the user and adds it to their browser and the internal cookie map
func cookieSetup (w http.ResponseWriter, r *http.Request, newUUID string, expDate time.Time) {
    //Generate & set browser cookie
    cookie := http.Cookie{Name: "localhost", Value: newUUID, Expires: expDate, HttpOnly: true, MaxAge: 100000, Path: "/"}
    http.SetCookie(w,&cookie)

    path := *templatesPath + "login.html"
    newTemplate,err := template.ParseFiles(path)   
    if err != nil {
	fmt.Println("Error running login template")
        if printToFile == 1 {
		defer Log.Flush()
    		Log.Error("Error running login template\r\n")
	}
	return;
    } 
    newTemplate.Execute(w,"loginTemplate")

    r.ParseForm()
    name := r.PostFormValue("name")
    submit := r.PostFormValue("submit") 

    if submit == "Submit" { // check if the user hit the "submit" button
    	if name == "" {
		path = *templatesPath + "/badLogin.html"
    		newTemplate,_ := template.New("outputUpdate").ParseFiles(path)   
    		newTemplate.ExecuteTemplate(w,"badLoginTemplate",nil)
    	} else {
		//generate cookie map's cookie
		mapCookie := http.Cookie{
		Name: newUUID, 
		Value: name, 
		Path: "/", 
		Domain: "localhost", 
		Expires: expDate,
 		HttpOnly: true, 
		MaxAge: 100000,
		}
		//lock the cookie map while it's being written to
		mapSetCookie(mapCookie, newUUID)

		fmt.Println("localhost:" + strconv.Itoa(*portNO) + "/login?name=" + name)
    		if  printToFile == 1 { //Check if the p2f flag is set
			defer Log.Flush()
    			Log.Info("localhost:" + strconv.Itoa(*portNO) + "/login?name=" + name + "\r\n")
    		}

		//Redirect to greetings (home) page
		path = *templatesPath + "greetingRedirect.html"
    		newTemp,err := template.New("redirect").ParseFiles(path)   
    		if err != nil {
			fmt.Println("Error running greeting redirect template")
        		if printToFile == 1 {
				defer Log.Flush()
    				Log.Error("Error running greeting redirect template\r\n")
			}
			return;
    		} 
    		newTemp.ExecuteTemplate(w,"greetingRedirectTemplate",portInfoStuff)
    	}
    }
}

//Retrieves the user name from their cookie on their browser (if one exists)
func getUserName (r *http.Request) string {
    for _, currCookie := range r.Cookies() { //Lookup the user name by cross matching the user cookie's value against the local cookie maps's cookie names
    	if (currCookie.Name != "") {
	currCookieVal := currCookie.Value
	mutex.Lock()
	mapCookie := cookieMap[currCookieVal]
	mutex.Unlock()
        	if (mapCookie.Value != "") {
    			return ", " + mapCookie.Value
		}
    	}
    }
    return ""
}