package database

import (
	"database/sql"
	"fmt"
	"myRoutine/models"
	"os"
	"io"
	"encoding/json"
	"log"
)

func ReadJsontoDB(fileDir string, db *sql.DB) {
	// Read all json data and update the DB
	jsonFile, err := os.Open(fileDir)
	if err != nil {
		fmt.Printf("error: %v \n", err)
		return
	}
	defer jsonFile.Close() // closing it right after the function ends.

	byteValue, _ := io.ReadAll(jsonFile)
	var sites models.SitesList

	err = json.Unmarshal(byteValue, &sites)
	if err != nil {
		fmt.Printf("error: %v \n", err)
		log.Fatal(err)
	}

	for i, site := range(sites.Websites) {
		// check if the same URL exists.
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM urls WHERE url = ?", site.Url).Scan(&count)
		if err != nil {
			fmt.Printf("error checking URL existence: %v\n", err)
			continue
		}
		// if this was a new URL, then add a row to the table of the URLs.
		// Url, Name are set according to the json but the status is false as a default.
		if count == 0 {
			_, err := db.Exec("INSERT INTO urls (url, name, status) VALUES (?, ?, ?)", site.Url, site.Name, false)
			if err != nil {
				fmt.Printf("error inserting URL: %v\n", err)
				continue
			}
		}
	}
} 

func ListSites(db *sql.DB) {
	// List all the sites in the data base.
}

func UpdateResponse(site models.site, db *sql.DB) {
	// Update the status of the website according to what it is
	// check if the previous response is the same as the new one first. maybe there is no need to update anything and make connections to db.
}

func AddSite(site models.site, db *sql.DB) {
	// Add a new site to the db
}

func RemoveSite(site models.site, db *sql.DB) {
	// Remove a row from the table
}

func UpdateUrl(site models.site, db *sql.DB) {
	// again like every update, make sure you are not making any unnecessary connections to DB. it costs alot.
}
