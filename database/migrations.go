package database

import (
    "database/sql"
    "fmt"
)

// CreateTables creates all necessary database tables
func CreateTables(db *sql.DB) error {
    // Create users table
    createUsersTable := `
    CREATE TABLE IF NOT EXISTS sites (
        name VARCHAR(100) NOT NULL,
        url VARCHAR(100) UNIQUE NOT NULL,
    	status BOOLEAN DEFAULT true
    )`
    
    _, err := db.Exec(createUsersTable)
    if err != nil {
        return fmt.Errorf("failed to create users table: %v", err)
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
    
    DROP TRIGGER IF EXISTS update_users_updated_at ON users;
    CREATE TRIGGER update_users_updated_at 
        BEFORE UPDATE ON users 
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    `
    
    _, err = db.Exec(createTrigger)
    if err != nil {
        return fmt.Errorf("failed to create trigger: %v", err)
    }
    
    return nil
}
