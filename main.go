package main

import (
	_ "embed"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	usernamePtr   *string
	passwordPtr   *string
	serverPtr     *string
	serverPortPtr *string
	branchPtr     *string
	publicPath    *string
	frontendPath  *string
	help          *bool
)

func displayHelp() {
	fmt.Printf("Usage of %s\n\n", color.Bold.Sprintf(os.Args[0]))
	fmt.Println("Upload a frontend folder to monkapps via SSH")
	fmt.Println("and generate an index to display which branches are available")
	fmt.Println("")
	fmt.Println("")
	color.Bold.Println("Credentials:")
	fmt.Println("  -username | Env variable[MONKAPPS_USERNAME]: Set the username for the connection to Monkapps.")
	fmt.Println("  -password | Env variable[MONKAPPS_PASSWORD]: Set the password for the connection to Monkapps.")
	fmt.Println("")
	color.Bold.Println("Connection:")
	fmt.Println("  -server | Env variable[MONKAPPS_SERVER]: Which server to connect to")
	fmt.Println("  -port | Env variable[MONKAPPS_SERVER_PORT]: Which tcp port to connect to")
	fmt.Println("")
	color.Bold.Println("Setup:")
	fmt.Println("  -publicPath | Env variable[MONKAPPS_FRONTEND_PATH]: What is the public path on monkapps defaults to ~/public when not provided.")
	fmt.Println("  -frontendPath | Env variable[MONKAPPS_FRONTEND_PATH]: What is the path with the compiled frontend code.")
	fmt.Println("")
	fmt.Println("")
	os.Exit(0)
}

func getBranchFromGit() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	stdout, err := cmd.Output()
	if err != nil {
		return ""
	}
	stdoutStr := fmt.Sprintf("%s", stdout)
	stdoutStr = strings.TrimRight(stdoutStr, "\n")
	return stdoutStr
}

func getVariableFromEnvironmentAndExit(key, error string) string {
	value := getVariableFromEnvironment(key)
	if len(value) == 0 {
		fmt.Printf("%s\n", error)
		os.Exit(1)
	}
	return value
}

func getVariableFromEnvironment(key string) string {
	value := os.Getenv(key)
	value = strings.TrimRight(value, "\n")
	return value
}

func overwriteVariableWithEnv(option *string, envVariable string, force bool) {
	if len(*option) > 0 && !force {
		return
	}
	if len(getVariableFromEnvironment(envVariable)) > 0 {
		*option = getVariableFromEnvironment(envVariable)
	}
}

func main() {
	var fileCount int = 0
	var totalBytes uint64 = 0
	usernamePtr = flag.String("username", "", "server username")
	passwordPtr = flag.String("password", "", "server password")
	serverPtr = flag.String("server", "eu.dev.monkapps.com", "Server")
	serverPortPtr = flag.String("port", "22", "Server port")
	branchPtr = flag.String("branch", "", "Git Branch")
	publicPath = flag.String("publicPath", "~/public", "Public path")
	frontendPath = flag.String("frontendPath", "", "Path to frontend")
	help = flag.Bool("help", false, "Display full help text.")
	flag.Parse()

	if *help {
		displayHelp()
	}

	overwriteVariableWithEnv(serverPtr, "MONKAPPS_SERVER", true)
	overwriteVariableWithEnv(serverPortPtr, "MONKAPPS_SERVER_PORT", true)
	overwriteVariableWithEnv(publicPath, "MONKAPPS_PUBLIC_PATH", true)
	overwriteVariableWithEnv(frontendPath, "MONKAPPS_FRONTEND_PATH", true)
	overwriteVariableWithEnv(usernamePtr, "MONKAPPS_USERNAME", false)
	overwriteVariableWithEnv(passwordPtr, "MONKAPPS_PASSWORD", false)

	if len(*branchPtr) == 0 {
		*branchPtr = getBranchFromGit()
		if len(*branchPtr) == 0 {
			*branchPtr = getVariableFromEnvironmentAndExit("MONKAPPS_BRANCH", "No branch supplied")
		}
	}

	if _, err := os.Stat(*frontendPath); os.IsNotExist(err) {
		fmt.Println("Couldn't find the frontend folder.")
		os.Exit(1)
	}

	config := &ssh.ClientConfig{
		User: *usernamePtr,
		Auth: []ssh.AuthMethod{
			ssh.Password(*passwordPtr),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", *serverPtr+":"+*serverPortPtr, config)
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		os.Exit(1)
	}

	defer func(conn *ssh.Client) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Could not close SSH connection. (%s)", err.Error())
		}
	}(conn)

	start := time.Now()

	// create new SFTP client
	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer func(client *sftp.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("Could not close SFTP connection. (%s)", err.Error())
		}
	}(client)

	branchRoot := addTrailingSlash(*publicPath) + addTrailingSlash(*branchPtr)

	var fileList []string
	err = filepath.Walk(*frontendPath, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})

	if err != nil {
		fmt.Println("Couldn't read files", err)
		os.Exit(1)
	}
	for _, path := range fileList {
		info, err := os.Stat(path)
		if err != nil {
			fmt.Println("Couldn't read file", err)
			os.Exit(1)
		}
		if !info.IsDir() {
			f, _ := os.Open(path)

			*branchPtr = replaceSlash(*branchPtr)

			dest := addTrailingSlash(*publicPath) + addTrailingSlash(*branchPtr) + path[len(*frontendPath)+1:]
			fmt.Print(path + " => " + dest)
			destDir := filepath.Dir(dest)
			client.MkdirAll(destDir)

			curTime := time.Now()
			client.Chtimes(destDir, curTime, curTime)

			// create destination file
			dstFile, err := client.Create(dest)
			if err != nil {
				fmt.Println("Error while creating file ", err)
				os.Exit(1)
			}

			// copy source file to destination file
			bytes, err := io.Copy(dstFile, f)
			if err != nil {
				fmt.Println("Error while copying file ", err)
				os.Exit(1)

			}
			fmt.Printf(" - %s copied\n", humanize.Bytes(uint64(bytes)))
			fileCount++
			totalBytes += uint64(bytes)
			dstFile.Close()
			f.Close()
		}
	}
	dstFile, err := client.Create(addTrailingSlash(*publicPath) + "index.php")
	if err != nil {
		fmt.Printf("Error while creating index.php file. (%s)\n", err.Error())
		os.Exit(1)
	}
	bytes, err := io.Copy(dstFile, indexPhpReader())
	fmt.Printf("Creating and copying new index.php - %s copied\n", humanize.Bytes(uint64(bytes)))
	fileCount++
	totalBytes += uint64(bytes)

	dstJson, err := client.Create(addTrailingSlash(branchRoot) + "deploy.json")
	if err != nil {
		fmt.Printf("Error while creating meta file. (%s)\n", err.Error())
		os.Exit(1)
	}
	bytes, err = io.Copy(dstJson, metadataJSON())
	fmt.Printf("Creating and copying new deploy.json - %s copied\n", humanize.Bytes(uint64(bytes)))
	fileCount++
	totalBytes += uint64(bytes)
	fmt.Printf("-----------------------------------------------------------------------------------\n")
	fmt.Printf("Files copied: %d (%s) - %s\n", fileCount, humanize.Bytes(totalBytes), time.Since(start))
	fmt.Printf("-----------------------------------------------------------------------------------\n\n")
}
