package database

import (
	"database/sql"
	"fmt"
	"myRoutine/models"
	"os"
	"io"
	"encoding/json"
)

func LoadFromJson(fileDir string) ([]models.Site, error) {
	jsonFile, err := os.Open(fileDir)
	if err != nil {
		fmt.Printf("error: %v \n", err)
		return nil, err
	}
	defer jsonFile.Close()
	
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("error reading file: %v \n", err)
		return nil, err
	}
	
	var sites models.SitesList
	err = json.Unmarshal(byteValue, &sites)
	if err != nil {
		fmt.Printf("error unmarshalling JSON: %v \n", err)
		return nil, err
	}
	
	return sites.Websites, nil
}

func ReadJsontoDB(fileDir string, db *sql.DB) error {
	// Read all JSON data and update the DB
	jsonFile, err := os.Open(fileDir)
	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		return err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("error reading file: %v\n", err)
		return err
	}

	var sites models.SitesList
	if err := json.Unmarshal(byteValue, &sites); err != nil {
		fmt.Printf("error unmarshalling JSON: %v\n", err)
		return err
	}

	for _, site := range sites.Websites {
		// check if the same URL exists
		var count int
		query := "SELECT COUNT(*) FROM urls WHERE url = $1"
		err := db.QueryRow(query, site.Url).Scan(&count)
		if err != nil {
			fmt.Printf("error checking URL existence: %v\n", err)
			continue
		}

		// if this is a new URL, insert it
		if count == 0 {
			_, err := db.Exec(
				"INSERT INTO urls (url, name, status) VALUES ($1, $2, $3)",
				site.Url, site.Name, false,
			)
			if err != nil {
				fmt.Printf("error inserting URL: %v\n", err)
				continue
			}
			fmt.Printf("Inserted new site: %s\n", site.Name)
		}
	}

	return nil
}

func ListSites(db *sql.DB) (models.SitesList, error) {
	// List all the sites in the database
	rows, err := db.Query("SELECT url, name, status FROM urls")
	if err != nil {
		return models.SitesList{}, fmt.Errorf("error querying sites: %v", err)
	}
	defer rows.Close()
	
	var sites models.SitesList
	for rows.Next() {
		var site models.Site
		if err := rows.Scan(&site.Url, &site.Name, &site.Status); err != nil {
			return models.SitesList{}, fmt.Errorf("error scanning site: %v", err)
		}
		sites.Websites = append(sites.Websites, site)
	}
	
	if err := rows.Err(); err != nil {
		return models.SitesList{}, fmt.Errorf("error iterating rows: %v", err)
	}
	
	return sites, nil
}

func UpdateResponse(site models.Site, db *sql.DB) error {
	// check if the previous response is the same as the new one first
	var currentStatus bool
	err := db.QueryRow("SELECT status FROM urls WHERE url = $1", site.Url).Scan(&currentStatus)
	if err != nil {
		return fmt.Errorf("failed to get current status: %v", err)
	}
	
	if currentStatus == site.Status {
		return nil // No change needed
	}
	
	// Update the status of the website
	res, err := db.Exec("UPDATE urls SET status = $1 WHERE url = $2", site.Status, site.Url)
	if err != nil {
		return fmt.Errorf("failed to update status: %v", err)
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	fmt.Printf("Status updated for %s: %t (%d rows affected)\n", site.Url, site.Status, rowsAffected)
	return nil
}

func AddSite(site models.Site, db *sql.DB) error {
	// Add a new site to the db
	_, err := db.Exec("INSERT INTO urls (url, name, status) VALUES ($1, $2, $3)", site.Url, site.Name, site.Status)
	if err != nil {
		return fmt.Errorf("could not insert the row: %v", err)
	}
	return nil
}

func RemoveSite(site models.Site, db *sql.DB) error {
	// Remove a row from the table
	_, err := db.Exec("DELETE FROM urls WHERE url = $1 OR name = $2", site.Url, site.Name)
	if err != nil {
		return fmt.Errorf("could not remove the row: %v", err)
	}
	return nil
}

func UpdateUrl(oldSite models.Site, newUrl string, db *sql.DB) error {
	res, err := db.Exec("UPDATE urls SET url = $1 WHERE url = $2 OR name = $3", newUrl, oldSite.Url, oldSite.Name)
	if err != nil {
		return fmt.Errorf("could not edit the row: %v", err)
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	fmt.Printf("URL updated successfully! %d rows affected\n", rowsAffected)
	return nil
}
