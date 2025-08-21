package main

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Site struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type SitesList struct {
	Websites []Site `json:"websites"`
}

func LoadFromJson(fileDir string) []Site {
	jsonFile, err := os.Open(fileDir)
	if err != nil {
		fmt.Printf("error: %v \n", err)
		log.Fatal(err)
	}
	defer jsonFile.Close() // closing it right after the function ends.

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var sites SitesList

	err = json.Unmarshal(byteValue, &sites)
	if err != nil {
		fmt.Printf("error: %v \n", err)
		log.Fatal(err)
	}
	return sites.Websites
}

func getStatus(url string, ch chan<- bool) {

	// fmt.Printf("checking %s\n", url)

	client := http.Client{
		Timeout: 30 * time.Second, // What is this and why is it defined like this? what is happening?
		}
	rsp, err := client.Get(url)
	if err != nil {
		// fmt.Printf("%s -> error: %v \n", url, err)
		ch <- false
		return
	}
	defer rsp.Body.Close() // What is the body? are we closing the client?
	ch <- rsp.StatusCode >= 200 && rsp.StatusCode < 400
	return
}

func updateStatus(url string, ch <-chan bool, done <-chan time.Time, runningStat chan<- bool) {
	fmt.Printf("open to receive form %s \n", url)
	for {
		select {
			case resp := <-ch:
				runningStat <- false
				if !resp {
					// update the list
					fmt.Printf("%s updated! --> %t\n" ,url, resp)
				}
			case <- done:
				return
		}
	}
}

func main() {

	jsonFile := "./websites.json"
	sites := LoadFromJson(jsonFile)

	// done := make(chan bool)
	ServerDone := time.After(20 * time.Minute) // using the server for 20 minuets and then turning it off.

	outputs := make([]chan bool, len(sites)) // a set of channels that each is declared statically // a set of channels that each is declared statically.
	running := make([]chan bool, len(sites)) // a set of channels that each is declared statically // a set of channels that each is declared statically.

	ticker := time.NewTicker(1 * time.Second) // how does this ticker thing work?
	defer ticker.Stop()

	for i, url := range(sites) {
		outputs[i] = make(chan bool)
		running[i] = make(chan bool, 1)
		running[i] <- false
		go updateStatus(url.Url, outputs[i], ServerDone, running[i])
	}

	for {
		select {
		case <- ticker.C:
			// run all go routines and check
			fmt.Printf("ticker ---------------\n")
			for i, url := range(sites) {
				select{
				case runningStat := <-running[i]:
					if !runningStat{
						go getStatus(url.Url, outputs[i])
						running[i] <- true
					} else {
						running[i] <- runningStat
					}
				}
			}
		case <- ServerDone:
			fmt.Printf("The job is Done. Bye Bye server. \n")
			return
		}
		// time.Sleep(1 * time.Second)
		// fmt.Printf("%v", ticker)
		
	}
}
