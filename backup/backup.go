package bp

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
)

var dumpDir string

func Backup() {
	// Open connection to database
	username := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")
	hostname := os.Getenv("DB_IP")
	port := os.Getenv("DB_PORT")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, hostname, port))
	log.Print("Success connecting to database")
	if err != nil {
		fmt.Println("Error opening database connection: ", err)

	}

	rows, err := db.Query("SHOW DATABASES")
	log.Print("success showing databases")
	if err != nil {
		fmt.Println("Error opening database: ", err)

	}

	// Iterate over the databases and perform backup for each one
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			fmt.Println("Error scanning database name: ", err)
		}

		// Skip system databases
		if dbName == "information_schema" || dbName == "mysql" || dbName == "performance_schema" || dbName == "sys" {
			continue
		}

		// Open connection to the specific database
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbName))
		if err != nil {
			fmt.Println("Error opening database connection: ", err)

		}
		dt := time.DateOnly
		// you should create this directory
		dumpFilenameFormat := fmt.Sprintf("%s-%s", dbName, dt)
		//localPath = dumpDir + "/" + dumpFilenameFormat
		dumpDir := os.Getenv("BACKUP_DIR")
		// Register database with mysqldump
		dumper, err := mysqldump.Register(db, dumpDir, dumpFilenameFormat)
		if err != nil {
			fmt.Println("Error registering database:", err)

		}

		// Dump database to file
		resultFilename, err := dumper.Dump()
		if err != nil {
			fmt.Println("Error dumping:", err)

		}
		log.Printf("File is saved to %s\n", resultFilename)

		// Upload the file to S3 bucket

		dumper.Close()
	}
	defer db.Close()
	defer rows.Close()
	//// Use the localPath variable outside the for loop
	//fmt.Printf("Local path: %s\n", localPath)
	err = uploadFile(dumpDir)
	log.Print("uploading...")
	if err != nil {
		fmt.Println("Error uploading file:", err)

	}
}
