package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func isolateDirectory(path string) error {
	if err := syscall.Chroot(path); err != nil {
		return err
	}

	if err := syscall.Chdir("/"); err != nil {
		return err
	}

	return nil
}

func copyExecutable(command string, path string) error {
	originalPath, err := os.Open(command)
	if err != nil {
		return err
	}

	newPath, err := os.OpenFile(filepath.Join(path, "executable"), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	_, err = io.Copy(newPath, originalPath)
	if err != nil {
		return err
	}

	err = originalPath.Close()
	if err != nil {
		return err
	}

	err = newPath.Close()
	if err != nil {
		return err
	}

	return nil
}

func runExecutable(args []string) {
	cmd := exec.Command("./executable", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{Cloneflags: syscall.CLONE_NEWPID}

	if err := cmd.Run(); err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
}

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	command := os.Args[3]
	args := os.Args[4:]

	tempPath, err := os.MkdirTemp("", "docker-fs")
	if err != nil {
		fmt.Printf("failed to create temporary dirpath: %v", err)
		os.Exit(1)
	}

	err = copyExecutable(command, tempPath)
	if err != nil {
		fmt.Printf("failed to copy executable: %v", err)
		os.Exit(1)
	}

	err = isolateDirectory(tempPath)
	if err != nil {
		fmt.Printf("failed to isolate directory: %v", err)
		os.Exit(1)
	}

	runExecutable(args)
}
