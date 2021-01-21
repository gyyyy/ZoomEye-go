package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

func winHome() string {
	var (
		drive = os.Getenv("HOMEDRIVE")
		path  = os.Getenv("HOMEPATH")
	)
	if drive == "" || path == "" {
		return os.Getenv("USERPROFILE")
	}
	return drive + path
}

func unixHome() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	var (
		out bytes.Buffer
		cmd = exec.Command("sh", "-c", "eval echo ~$USER")
	)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func home() string {
	if user, err := user.Current(); nil == err {
		return user.HomeDir
	}
	if runtime.GOOS == "windows" {
		return winHome()
	}
	return unixHome()
}

func checkFolder(dir *string) error {
	if strings.HasPrefix(*dir, "~/") {
		if base := home(); base != "" {
			*dir = filepath.Join(base, strings.TrimPrefix(*dir, "~/"))
		}
	}
	if info, err := os.Stat(*dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if info != nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(*dir, os.ModePerm)
}

func readFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func writeFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0o600)
}

func appendToFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}
