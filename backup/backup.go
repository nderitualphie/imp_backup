package bp

import (
	"database/sql"
	"fmt"
	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

func Backup() {
	// Open connection to database
	username := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")
	hostname := os.Getenv("DB_IP")
	port := os.Getenv("DB_PORT")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, hostname, port))
	if err != nil {
		fmt.Println("Error opening database connection: ", err)
		return
	}
	defer db.Close()
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return
	}
	defer rows.Close()

	// Declare localPath variable
	var localPath string

	// Iterate over the databases and perform backup for each one
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			fmt.Println("Error scanning database name: ", err)
			return
		}

		// Skip system databases
		if dbName == "information_schema" || dbName == "mysql" || dbName == "performance_schema" || dbName == "sys" {
			continue
		}

		// Open connection to the specific database
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbName))
		if err != nil {
			fmt.Println("Error opening database connection: ", err)
			return
		}
		dumpDir := os.Getenv("BACKUP_DIR") // you should create this directory
		dumpFilenameFormat := fmt.Sprintf("%s-%s.sql", dbName, time.Now().Format("20060102"))
		localPath = dumpDir + "/" + dumpFilenameFormat

		// Register database with mysqldump
		dumper, err := mysqldump.Register(db, dumpDir, dumpFilenameFormat)
		if err != nil {
			fmt.Println("Error registering database:", err)
			return
		}

		// Dump database to file
		resultFilename, err := dumper.Dump()
		if err != nil {
			fmt.Println("Error dumping:", err)
			return
		}
		fmt.Printf("File is saved to %s\n", resultFilename)

		// Close dumper and connected database
		dumper.Close()

		// Upload the file to S3 bucket
		bucketName := os.Getenv("BUCKET_NAME")
		remotePath := dumpFilenameFormat
		err = UploadObject(localPath, bucketName, remotePath)
		if err != nil {
			log.Printf("Error uploading to bucket: %v", err)
			return
		}
	}

	// Use the localPath variable outside the for loop
	fmt.Printf("Local path: %s\n", localPath)

}
