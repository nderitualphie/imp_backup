package main

import (
	bp "backup/backup"
	"fmt"
)

func main() {
	// Call the Backup function to perform the database backup
	bp.Backup()

	fmt.Println("Database backup completed.")
}
