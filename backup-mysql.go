package main

import (
	"os"
	"os/exec"
	"fmt"
	"strings"
	"time"
	"io/ioutil"
)

func main() {
	// Command-line argument check
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  To initialize a configuration file: ./backup-mysql init <file>")
		fmt.Println("  To perform MySQL backup: ./backup-mysql export <file>")
		fmt.Println("  To execute rclone sync: ./backup-mysql rclone <file>")
		return
	}

	command := os.Args[1]
	configFile := os.Args[2]

	switch command {
	case "init":
		initializeConfig(configFile)
	case "export":
		exportMySQL(configFile)
	case "rclone":
		runRclone(configFile)		
	default:
		fmt.Println("Invalid command. Please use 'init', 'export', or 'rclone'.")
	}
}

func initializeConfig(fileName string) {
	configContent := `[client]
host=
port=
user=
password=

[database]
dbname=

[rclone]
backupdir=
rcloneremotes=
clouddir=`

	err := ioutil.WriteFile(fileName, []byte(configContent), 0600)
	if err != nil {
		fmt.Println("Error creating configuration file:", err)
		return
	}

	fmt.Println("Configuration file created:", fileName)
}

func exportMySQL(configFile string) {
	// Read configuration file
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error reading configuration file:", err)
		return
	}

	// Parse configuration file
	configLines := strings.Split(string(content), "\n")
	var dbName, backupDir string

	for _, line := range configLines {
		if strings.HasPrefix(line, "dbname=") {
			dbName = strings.TrimPrefix(line, "dbname=")
		} else if strings.HasPrefix(line, "backupdir=") {
			backupDir = strings.TrimPrefix(line, "backupdir=")
		}
	}

	currentDate := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("%s-%s.sql.gz", dbName, currentDate)

	// Check if the backup directory exists, if not, create it
	_, err = os.Stat(backupDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(backupDir, 0755)
		if err != nil {
			fmt.Println("Error creating backup directory:", err)
			return
		}
	}

	backupFilePath := fmt.Sprintf("%s/%s", backupDir, backupFileName)

	// Using native gzip for compression
	cmd := exec.Command("sh", "-c", fmt.Sprintf("mysqldump --defaults-extra-file=%s %s | gzip > %s", configFile, dbName, backupFilePath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %s\nOutput: %s\n", err, string(output))
		return
	}

	fmt.Println("Backup completed:", backupFileName)
}

func runRclone(configFile string) {
	// Read configuration file and parse rclone remotes
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error reading configuration file:", err)
		return
	}

	configLines := strings.Split(string(content), "\n")
	var backupDir, rcloneRemotes, rcloneDir string

	for _, line := range configLines {
		if strings.HasPrefix(line, "backupdir=") {
			backupDir = strings.TrimPrefix(line, "backupdir=")
		} else if strings.HasPrefix(line, "rcloneremotes=") {
			rcloneRemotes = strings.TrimPrefix(line, "rcloneremotes=")
		} else if strings.HasPrefix(line, "clouddir=") {
			rcloneDir = strings.TrimPrefix(line, "clouddir=")
		}
	}

	if rcloneRemotes == "" {
		fmt.Println("Error: rcloneremotes is not specified in the configuration file.")
		return
	}

	// Perform rclone commands
	rcloneRemotesList := strings.Split(rcloneRemotes, ", ")

	for _, remote := range rcloneRemotesList {
		rcloneCommand := exec.Command("rclone", "sync", "-v", backupDir, remote+":"+rcloneDir)
		rcloneCommand.Stdout = os.Stdout
		rcloneCommand.Stderr = os.Stderr

		err := rcloneCommand.Run()
		if err != nil {
			fmt.Println("Error running rclone command:", err)
			return
		}
	}

	fmt.Println("rclone completed for:", backupDir)
}