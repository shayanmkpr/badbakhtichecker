// check if a site goes down or comes back up. toggle status and then give a notification.
package main

import (
	"fmt"
	"time"
	"net/http"
	"log"

	"myRoutine/database"
	"myRoutine/config"
	"myRoutine/models"
	"database/sql"
	_ "github.com/lib/pq"
)

/*
the routes:
read the json file and import everysite --> DB --> list sites for the main.go --> {/ticker/ -->check the status of all --> update DB}
											|--> Update DB with new sites --|
											|--> Remove sites --|
											|--> Add sites --|
*/

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

func updateStatus(db *sql.DB, site models.Site,
	ch <-chan bool, done <-chan time.Time, runningStat *models.Runner) {

	fmt.Printf("open to receive form %s \n", site.Url)
	for {
		select {
			case resp := <-ch:
				runningStat.Mu.Lock()
				runningStat.Running[site.Url] = false
				runningStat.Mu.Unlock()
				// update the corresponding db row with the new stat
				database.UpdateResponse(models.Site{Url: site.Url, Name: site.Name, Status: resp}, db)
			case <- done:
				return
		}
	}
}

func main() {


	// Connect to PostgreSQL
	db, err := config.ConnectDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
	fmt.Printf("the connection was established \n")

	//Create the DB if it doesnt exist. The if is being handled in the function.
	error := database.CreateTables(db)
	if error != nil {
		fmt.Printf("could not create tables: %v", err)
		log.Fatal(err)
	}

	fmt.Printf("the tables were created or they were already there idk \n")

	jsonFile := "./websites.json"
	err = database.ReadJsontoDB(jsonFile, db) // read from json and put them into the DB
	if err != nil{
		fmt.Printf("the ReadJsontoDB error %v \n", err)
	}
	sites, err := database.ListSites(db) // load sites list form DB
	if err != nil {
		fmt.Printf("could not list sites from DB: %v", err)
		log.Fatal(err)
	}

	ServerDone := time.After(20 * time.Minute) // using the server for 20 minuets and then turning it off.

	outputs := make([]chan bool, len(sites.Websites)) // a set of channels that each is declared statically // a set of channels that each is declared statically.
	runningStat := make([]*models.Runner, len(sites.Websites))

	ticker := time.NewTicker(1 * time.Second) // how does this ticker thing work?
	defer ticker.Stop()

	for i, site := range(sites.Websites) {
		outputs[i] = make(chan bool, 1)
		runningStat[i] = &models.Runner{Running: make(map[string] bool)} // initializing the runningStat memory
		runningStat[i].Mu.Lock()
		runningStat[i].Running[site.Url] = false
		runningStat[i].Mu.Unlock()
		go updateStatus(db, site, outputs[i], ServerDone, runningStat[i])
	}

	// here we are checking and updating everything for real. this is the main ticker loop.
	for {
		select {
		case <- ticker.C:
			// run all go routines and check
			fmt.Printf("ticker ---------------\n")
			for i, url := range(sites.Websites) {
				if runningStat[i].Running[url.Url] == false {
					go getStatus(url.Url, outputs[i])
					runningStat[i].Mu.Lock()
					runningStat[i].Running[url.Url] = true
					runningStat[i].Mu.Unlock()
				} else {
					fmt.Printf("%s is still Running\n", url.Url)
				}
			}
		case <- ServerDone:
			fmt.Printf("The job is Done. Bye Bye server. \n")
			return
		}
	}
}
