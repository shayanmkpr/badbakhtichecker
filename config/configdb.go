package config
import (
    "database/sql"
    "fmt"
)

func CreateTables(db *sql.DB) error {
    createSitesTable := `
    CREATE TABLE IF NOT EXISTS sites (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        url VARCHAR(100) UNIQUE NOT NULL,
        status BOOLEAN DEFAULT true,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`
    
    _, err := db.Exec(createSitesTable)
    if err != nil {
        fmt.Printf("error: %v \n", err)
        return fmt.Errorf("failed to create sites table: %v", err)
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
    
    DROP TRIGGER IF EXISTS update_sites_updated_at ON sites;
    CREATE TRIGGER update_sites_updated_at 
        BEFORE UPDATE ON sites 
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    `
    
    _, err = db.Exec(createTrigger)
    if err != nil {
        fmt.Printf("error: %v \n", err)
        return fmt.Errorf("failed to create trigger: %v", err)
    }
    
    return nil
}
