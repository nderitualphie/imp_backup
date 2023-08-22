package bp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func Backup() {
	// Get container name and database name as environment variables
	containerName := os.Getenv("C_NAME")
	databaseName := os.Getenv("DB_NAME")

	// Create a timestamp-based filename for the backup
	timestamp := time.Now().Format("2006-01-02")
	backupFileName := fmt.Sprintf("%s_%s.sql", timestamp, databaseName)

	// Create a directory for backups if it doesn't exist
	backupDir := os.Getenv("BACKUP_DIR")
pass := os.Getenv("DB_PASSWORD")
uname := os.Getenv("DB_USER")
	// Build the mysqldump command
	cmd := exec.Command(
    "docker", "exec", containerName,
    "mysqldump", "-u", uname, "--password="+pass, databaseName,
    ">", fmt.Sprintf("%s/%s", backupDir, backupFileName),

)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running mysqldump: %v\n", err)
		log.Printf("Command output: %s\n", output)
		return
	}

	// Save the backup to a file
	backupFilePath := fmt.Sprintf("%s/%s", backupDir, backupFileName)
	err = os.WriteFile(backupFilePath, output, 0644)
	if err != nil {
		log.Printf("Error saving backup file: %v\n", err)
		return
	}

	// Upload the backup file
	err = uploadFile(backupDir)
	if err != nil {
		log.Printf("Error uploading file: %v\n", err)
		return
	}
	log.Print("Upload completed successfully")
}
