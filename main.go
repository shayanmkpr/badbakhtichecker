// check if a site goes down or comes back up. toggle status and then give a notification.
package main

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type Site struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type SitesList struct {
	Websites []Site `json:"websites"`
}

type Runner struct {
	mu sync.Mutex
	running map[string] bool
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

func updateStatus(url string, ch <-chan bool, done <-chan time.Time, runningStat *Runner) {
	fmt.Printf("open to receive form %s \n", url)
	for {
		select {
			case resp := <-ch:
				runningStat.mu.Lock()
				runningStat.running[url] = false
				runningStat.mu.Unlock()
				if !resp{
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

	ServerDone := time.After(20 * time.Minute) // using the server for 20 minuets and then turning it off.

	outputs := make([]chan bool, len(sites)) // a set of channels that each is declared statically // a set of channels that each is declared statically.
	runningStat := make([]*Runner, len(sites))

	ticker := time.NewTicker(1 * time.Second) // how does this ticker thing work?
	defer ticker.Stop()

	for i, url := range(sites) {
		outputs[i] = make(chan bool, 1)
		runningStat[i] = &Runner{running: make(map[string] bool)} // initializing the runningStat memory
		runningStat[i].mu.Lock()
		runningStat[i].running[url.Url] = false
		runningStat[i].mu.Unlock()
		go updateStatus(url.Url, outputs[i], ServerDone, runningStat[i])
	}

	for {
		select {
		case <- ticker.C:
			// run all go routines and check
			fmt.Printf("ticker ---------------\n")
			for i, url := range(sites) {
				if runningStat[i].running[url.Url] == false {
					go getStatus(url.Url, outputs[i])
					runningStat[i].mu.Lock()
					runningStat[i].running[url.Url] = true
					runningStat[i].mu.Unlock()
				} else {
					fmt.Printf("%s is still running\n", url.Url)
				}
			}
		case <- ServerDone:
			fmt.Printf("The job is Done. Bye Bye server. \n")
			return
		}
	}
}
