package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// Debugging print statements
	// fmt.Println("Logs from your program will appear here!")

	command := os.Args[3]
	args := os.Args[4:]

	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	original_path, err := os.Open(command)
	if err != nil {
		fmt.Println("failed to open command binary:", err.Error())
		os.Exit(1)
	}

	if err = os.Mkdir("/docker-fs", 0755); err != nil {
		fmt.Println("failed to mkdir:", err.Error())
		os.Exit(1)
	}

	if err = syscall.Chroot("/docker-fs"); err != nil {
		fmt.Println("failed to chroot:", err.Error())
		os.Exit(1)
	}

	if err = syscall.Chdir("/"); err != nil {
		fmt.Println("failed to chdir:", err.Error())
		os.Exit(1)
	}

	new_path, err := os.OpenFile("executable", os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println("failed to open copy location:", err.Error())
		os.Exit(1)
	}

	_, err = io.Copy(new_path, original_path)
	if err != nil {
		fmt.Println("failed to copy file:", err.Error())
		os.Exit(1)
	}

	if err = cmd.Run(); err != nil {
		fmt.Println("Err:", err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
