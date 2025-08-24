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

func LoadFromJson(fileDir string) ([]models.Site, error) {
	jsonFile, err := os.Open(fileDir)
	if err != nil {
		fmt.Printf("error: %v \n", err)
		log.Fatal(err)
	}
	defer jsonFile.Close() // closing it right after the function ends.

	byteValue, _ := io.ReadAll(jsonFile)
	var sites models.SitesList

	err = json.Unmarshal(byteValue, &sites)
	if err != nil {
		fmt.Printf("error: %v \n", err)
		log.Fatal(err)
	}
	return sites.Websites, nil
}

func ReadJsontoDB(fileDir string, db *sql.DB) error {
	// Read all json data and update the DB
	jsonFile, err := os.Open(fileDir)
	if err != nil {
		return fmt.Printf("error: %v \n", err)
	}
	defer jsonFile.Close() // closing it right after the function ends.

	byteValue, _ := io.ReadAll(jsonFile)
	var sites models.SitesList

	err = json.Unmarshal(byteValue, &sites)
	if err != nil {
		fmt.Printf("error: %v \n", err)
		log.Fatal(err)
	}

	for _, site := range(sites.Websites) {
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

func ListSites(db *sql.DB) (models.SiteList, error){
	// List all the sites in the data base if they have complete info stored.
	rows, err := db.Query("SELECT url, name, stats FROM urls")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var sites models.SitesList
	for rows.Next() {
		var site models.Sit&e
		if err := rows.Scan(&site.Url, &site.Name, &site.Status); err != nil {
			log.Fatal(err)
		}
		sites.Websites = append(sites.Websites, site)
	}

	if err := rows.Err(); err != nil {
		return models.SitesList{}, err
	}
	return (sites, nil)
}

func UpdateResponse(site models.Site, db *sql.DB) error {

	// check if the previous response is the same as the new one first. maybe there is no need to update anything and make connections to db.
	var currentStatus bool
	db.QueryRow("SELECT stats FROM urls WHERE url = ?", site.Status).Scan(&currentStatus)

	if currentStatus == site.Status {
		return nil
	}

	// Update the status of the website according to what it is
	res, err := db.Exec("Update urls set stats = ? WHERE url = ?", site.Status, site.Url)
	if err != nil {
		return fmt.Printf("Error: failed to get current status: %v", err)
    }

	rowsAffected := res.RowsAffected()
	fmt.printf("%d rows updated", rowsAffected)
	return nil
}

func AddSite(site models.Site, db *sql.DB) error {
	// Add a new site to the db
	_, err := db.Exec("INSERT INTO urls (url, name, status) VALUES (?, ?, ?)", site.Url, site.Name, site.Status)
	if err != nil {
		log.Fatal(err)
		return fmt.Printf("Could not insert the row. %v", err)
	}
	return nil
}

func RemoveSite(site models.Site, db *sql.DB) error {
	// Remove a row from the table
	_, err := db.Exec("DELETE FROM urls WHERE url = ? OR name = ?", site.Url, site.Name)
	if err != nil {
		log.Fatal(err)
		return fmt.Printf("Could not remove the row. %v", err)
	}
	return nil
}

func UpdateUrl(site models.Site, db *sql.DB) {
	// again like every update, make sure you are not making any unnecessary connections to DB. it costs alot.
	res, err := db.Exec("Update urls set url = ? WHERE url = ? OR name = ?")
	if err != nil {
		log.Fatal(err)
		return fmt.Printf("Could not edit the row. %v", err)
	}
	updatedRow := res
	fmt.Printf("row was updated successfully! %v", res)
	return nil
}
