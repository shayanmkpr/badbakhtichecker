package config
import (
    "database/sql"
    "fmt"
)

// CreateTables creates the required tables if they don't exist
func CreateTables(db *sql.DB) error {
	createUrlsTable := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		url VARCHAR(255) UNIQUE NOT NULL,
		status BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	_, err := db.Exec(createUrlsTable)
	if err != nil {
		fmt.Printf("error creating urls table: %v \n", err)
		return fmt.Errorf("failed to create urls table: %v", err)
	}
	
	// Create trigger to automatically update 'updated_at' field
	createTrigger := `
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ language 'plpgsql';
	
	DROP TRIGGER IF EXISTS update_urls_updated_at ON urls;
	CREATE TRIGGER update_urls_updated_at 
		BEFORE UPDATE ON urls 
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
	`
	
	_, err = db.Exec(createTrigger)
	if err != nil {
		fmt.Printf("error creating trigger: %v \n", err)
		return fmt.Errorf("failed to create trigger: %v", err)
	}
	
	return nil
}
