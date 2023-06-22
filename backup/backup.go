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
	dbname := os.Getenv("DB_NAME")

	dumpDir := os.Getenv("BACKUP_DIR") // you should create this directory
	dt := time.DateTime
	dumpFilenameFormat := fmt.Sprintf("%v:%v", dbname, dt)

	// accepts time layout string and add .sql at the end of file
	localPath := dumpDir + "/" + dumpFilenameFormat
	Bucket := os.Getenv("BUCKET_NAME")
	remotePath := ""
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname))
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return
	}

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
	fmt.Printf("File is saved to %s", resultFilename)

	// Close dumper and connected database
	dumper.Close()
	err = UploadObject(localPath, Bucket, remotePath)
	if err != nil {
		log.Printf("Couldnt Uploaad: %v", err)
	}
}
