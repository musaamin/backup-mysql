package main

import (
	"os"
	"os/exec"
	"fmt"
	"strings"
	"time"
	"io/ioutil"
	"compress/gzip"
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
	configContent := `DBHOST=
DBPORT=
DBNAME=
DBUSER=
DBPASS=
BACKUPDIR=
RCLONEREMOTES=
RCLONEDIR=`

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
	var dbHost, dbPort, dbName, dbUser, dbPass, backupDir string

	for _, line := range configLines {
		if strings.HasPrefix(line, "DBHOST=") {
			dbHost = strings.TrimPrefix(line, "DBHOST=")
		} else if strings.HasPrefix(line, "DBPORT=") {
			dbPort = strings.TrimPrefix(line, "DBPORT=")
		} else if strings.HasPrefix(line, "DBNAME=") {
			dbName = strings.TrimPrefix(line, "DBNAME=")
		} else if strings.HasPrefix(line, "DBUSER=") {
			dbUser = strings.TrimPrefix(line, "DBUSER=")
		} else if strings.HasPrefix(line, "DBPASS=") {
			dbPass = strings.TrimPrefix(line, "DBPASS=")
		} else if strings.HasPrefix(line, "BACKUPDIR=") {
			backupDir = strings.TrimPrefix(line, "BACKUPDIR=")
		}
	}

	currentDate := time.Now().Format("20060102")
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

	cmd := exec.Command("mysqldump", "-h"+dbHost, "-P"+dbPort, "-u"+dbUser, "-p"+dbPass, dbName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		var errMsg string
		if exitError, ok := err.(*exec.ExitError); ok {
			exitStatus := exitError.ExitCode()
			if exitStatus == 2 {
				errMsg = "Error: Authentication Failed or Database Does Not Exist."
			} else {
				errMsg = fmt.Sprintf("Error running mysqldump: Exit Status %d", exitStatus)
			}
		} else {
			errMsg = "Error running mysqldump: " + err.Error()
		}
		fmt.Println(errMsg)
		return
	}

	outFile, err := os.Create(backupDir + "/" + backupFileName)
	if err != nil {
		fmt.Println("Error creating backup file:", err)
		return
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	_, err = gzWriter.Write(output)
	if err != nil {
		fmt.Println("Error writing backup content:", err)
		return
	}
	defer gzWriter.Close()

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
		if strings.HasPrefix(line, "BACKUPDIR=") {
			backupDir = strings.TrimPrefix(line, "BACKUPDIR=")
		} else if strings.HasPrefix(line, "RCLONEREMOTES=") {
			rcloneRemotes = strings.TrimPrefix(line, "RCLONEREMOTES=")
		} else if strings.HasPrefix(line, "RCLONEDIR=") {
			rcloneDir = strings.TrimPrefix(line, "RCLONEDIR=")
		}
	}

	if rcloneRemotes == "" {
		fmt.Println("Error: RCLONEREMOTES is not specified in the configuration file.")
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