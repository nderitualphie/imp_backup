package bp

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"os"
	"path/filepath"
	"time"
)

func Backup() {
	// Get container name and database name as command-line arguments
	containerName := os.Getenv("C_NAME")
	databaseName := os.Getenv("DB_NAME")

	// Create a timestamp-based filename for the backup
	timestamp := time.DateOnly
	backupFileName := fmt.Sprintf("%s_%s.sql", timestamp, databaseName)

	// Create a directory for backups if it doesn't exist
	backupDir := os.Getenv("BACKUP_DIR")

	// Initialize Docker client
	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Println("Error initializing Docker client:", err)

	}
	uname := os.Getenv("")
	// Run backup command inside the container
	cmd := []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf("mysqldump -u<username> -p<password> %s > /backup/%s", databaseName, backupFileName),
	}
	createResp, err := cli.ContainerExecCreate(context.Background(), containerName, types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		fmt.Println("Error creating exec instance:", err)
		return
	}

	resp, err := cli.ContainerExecAttach(context.Background(), createResp.ID, types.ExecStartCheck{})
	if err != nil {
		fmt.Println("Error attaching to exec instance:", err)
		return
	}
	defer resp.Close()

	// Copy the backup file from the container to the host
	destPath := filepath.Join(backupDir, backupFileName)
	out, err := os.Create(destPath)
	if err != nil {
		fmt.Println("Error creating backup file:", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Reader)
	if err != nil {
		fmt.Println("Error copying backup data:", err)
		return
	}

	fmt.Printf("Database backup saved to %s\n", destPath)
}
