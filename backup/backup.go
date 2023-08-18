package bp

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"os"
	"os/exec"
	"time"
)

func Backup() {
	// Get container name and database name as command-line arguments
	containerName := os.Getenv("C_NAME")
	databaseName := os.Getenv("DB_NAME")

	// Create a timestamp-based filename for the backup
	timestamp := time.Now().Format("2006-01-02")
	backupFileName := fmt.Sprintf("%s_%s.sql", timestamp, databaseName)

	// Create a directory for backups if it doesn't exist
	backupDir := os.Getenv("BACKUP_DIR")

	// Initialize Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println("Error initializing Docker client:", err)

	}
	uname := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	// Run backup command inside the container

	cmd := []string{
		"/bin/sh",
		"-c",

		fmt.Sprintf("mysqldump -u %s -p'%s' %s > %s", uname, pass, databaseName, backupFileName),
	}

	createResp, err := cli.ContainerExecCreate(context.Background(), containerName, types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		fmt.Println("Error creating exec instance:", err)

	}

	resp, err := cli.ContainerExecAttach(context.Background(), createResp.ID, types.ExecStartCheck{})
	if err != nil {
		fmt.Println("Error attaching to exec instance:", err)

	}
	defer resp.Close()
	copyCmd := exec.Command("docker", "cp", containerName+":"+backupFileName, backupDir)
	copyOutput, copyErr := copyCmd.CombinedOutput()
	if copyErr != nil {
		log.Printf("Error copying backup file from container: %v\n", copyErr)
		log.Printf("docker cp output: %s\n", copyOutput)
	}
	err = uploadFile(backupDir)
	log.Print("uploading...")
	if err != nil {
		fmt.Println("Error uploading file:", err)

	}
}
