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

func getStatus(url string, ch chan<- bool) {
	fmt.Printf("Checking %s\n", url)
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	
	rsp, err := client.Get(url)
	if err != nil {
		fmt.Printf("%s -> error: %v\n", url, err)
		ch <- false
		return
	}
	defer rsp.Body.Close()
	
	isUp := rsp.StatusCode >= 200 && rsp.StatusCode < 400
	fmt.Printf("%s -> Status: %d (Up: %t)\n", url, rsp.StatusCode, isUp)
	ch <- isUp
}

func updateStatus(db *sql.DB, site models.Site, ch <-chan bool, done <-chan time.Time, runningStat *models.Runner) {
	fmt.Printf("Listening for updates from %s\n", site.Url)
	for {
		select {
		case resp := <-ch:
			runningStat.Mu.Lock()
			runningStat.Running[site.Url] = false
			runningStat.Mu.Unlock()
			
			// Update the corresponding db row with the new status
			err := database.UpdateResponse(models.Site{Url: site.Url, Name: site.Name, Status: resp}, db)
			if err != nil {
				fmt.Printf("Error updating status for %s: %v\n", site.Url, err)
			}
		case <-done:
			fmt.Printf("Stopping updates for %s\n", site.Url)
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
	fmt.Printf("Database connection established\n")

	// Create the tables if they don't exist
	// Note: Make sure you're using the CreateTables function from the database package
	err = config.CreateTables(db)
	if err != nil {
		fmt.Printf("Could not create tables: %v\n", err)
		log.Fatal(err)
	}
	fmt.Printf("Tables created or verified\n")

	// Read from JSON and populate the DB
	jsonFile := "./websites.json"
	err = database.ReadJsontoDB(jsonFile, db)
	if err != nil {
		fmt.Printf("ReadJsontoDB error: %v\n", err)
		log.Fatal(err)
	}
	fmt.Printf("JSON data loaded to DB\n")

	// Load sites list from DB
	sites, err := database.ListSites(db)
	if err != nil {
		fmt.Printf("Could not list sites from DB: %v\n", err)
		log.Fatal(err)
	}
	fmt.Printf("Loaded %d sites from DB\n", len(sites.Websites))

	if len(sites.Websites) == 0 {
		fmt.Printf("No sites found in database. Make sure your JSON file exists and contains valid data.\n")
		return
	}

	// Server will run for 20 minutes
	serverDone := time.After(20 * time.Minute)
	
	// Create channels and running status for each site
	outputs := make([]chan bool, len(sites.Websites))
	runningStat := make([]*models.Runner, len(sites.Websites))
	
	// Ticker to check sites every 30 seconds (reduced from 1 second to be less aggressive)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initialize channels and running status for each site
	for i, site := range sites.Websites {
		outputs[i] = make(chan bool, 1)
		runningStat[i] = &models.Runner{Running: make(map[string]bool)}
		runningStat[i].Mu.Lock()
		runningStat[i].Running[site.Url] = false
		runningStat[i].Mu.Unlock()
		
		// Start the update goroutine for this site
		go updateStatus(db, site, outputs[i], serverDone, runningStat[i])
	}

	fmt.Printf("Starting monitoring of %d sites...\n", len(sites.Websites))

	// Main monitoring loop
	for {
		select {
		case <-ticker.C:
			fmt.Printf("=== Checking all sites ===\n")
			for i, site := range sites.Websites {
				runningStat[i].Mu.Lock()
				isRunning := runningStat[i].Running[site.Url]
				runningStat[i].Mu.Unlock()
				
				if !isRunning {
					runningStat[i].Mu.Lock()
					runningStat[i].Running[site.Url] = true
					runningStat[i].Mu.Unlock()
					
					go getStatus(site.Url, outputs[i])
				} else {
					fmt.Printf("%s is still being checked\n", site.Url)
				}
			}
		case <-serverDone:
			fmt.Printf("Monitoring complete. Shutting down.\n")
			return
		}
	}
}
