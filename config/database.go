package config

import (
	"database/sql"
	"fmt"
	-"github.com/lib/pq"
)

type DatabaseConfig struct {
	Host     string
	Port 	 int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:	"localhost",
		Port:	5432,
		User:	"shayan", // should be set using env variables
		Password: "shayanmkpr", //should be set using env varialbes
		DBName:   "health", //should be set using env variables
		SSLMode:  "disable"
	}
}

// ConnectDB creates and returns a database connection
func ConnectDB() (*sql.DB, error) {
    config := GetDatabaseConfig()
    
    // Create connection string
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
    
    // Open database connection
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %v", err)
    }
    
    // Test the connection
    if err = db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %v", err)
    }
    
    return db, nil
}
