package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

var appURL string

func doNew(appName string) {
	appName = strings.ToLower(appName)
	appURL = appName

	// sanitize the application name (convert url to single word)
	if strings.Contains(appName, "/") {
		exploded := strings.SplitAfter(appName, "/")
		appName = exploded[(len(exploded) - 1)]
	}

	log.Println("App name is", appName)

	// git clone the skeleton application
	color.Green("\tCloning repository...")
	url := "git@github.com:yamagit01/celeritas-app.git"
	_, err := git.PlainClone(filepath.Join(".", appName), false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		exitGracefully(err)
	}

	// remove .git directory
	err = os.RemoveAll(filepath.Join(".", appName, ".git"))
	if err != nil {
		exitGracefully(err)
	}

	// create a ready to go .env file
	color.Yellow("\tCreating .env file...")
	data, err := templateFS.ReadFile(filepath.Join("templates", "env.txt"))
	if err != nil {
		exitGracefully(err)
	}

	env := string(data)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	env = strings.ReplaceAll(env, "${KEY}", cel.RandomString(32))

	err = copyDataToFile([]byte(env), filepath.Join(".", appName, ".env"))
	if err != nil {
		exitGracefully(err)
	}

	// create a makefile
	if runtime.GOOS == "windows" {
		source, err := os.Open(filepath.Join(".", appName, "Makefile.windows"))
		if err != nil {
			exitGracefully(err)
		}
		defer source.Close()

		destination, err := os.Create(filepath.Join(".", appName, "Makefile"))
		if err != nil {
			exitGracefully(err)
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		if err != nil {
			exitGracefully(err)
		}
	} else {
		source, err := os.Open(filepath.Join(".", appName, "Makefile.mac"))
		if err != nil {
			exitGracefully(err)
		}
		defer source.Close()

		destination, err := os.Create(filepath.Join(".", appName, "Makefile"))
		if err != nil {
			exitGracefully(err)
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		if err != nil {
			exitGracefully(err)
		}
	}
	_ = os.Remove(filepath.Join(".", appName, "Makefile.mac"))
	_ = os.Remove(filepath.Join(".", appName, "Makefile.windows"))

	// update the go.mod file
	color.Yellow("\tCreating go.mod file...")
	_ = os.Remove(filepath.Join(".", appName, "go.mod"))

	data, err = templateFS.ReadFile(filepath.Join("templates", "go.mod.txt"))
	if err != nil {
		exitGracefully(err)
	}

	mod := string(data)
	mod = strings.ReplaceAll(mod, "${APP_NAME}", appURL)

	err = copyDataToFile([]byte(mod), filepath.Join(".", appName, "go.mod"))
	if err != nil {
		exitGracefully(err)
	}

	// update existing .go files with correct name/imports
	color.Yellow("\tUpdating source files...")
	os.Chdir(filepath.Join(".", appName))
	updateSource()

	// run go mod tidy in the project directory
	color.Yellow("\tRunning go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	err = cmd.Start()
	if err != nil {
		exitGracefully(err)
	}

	color.Green("Done building" + appURL)
	color.Green("Go build something awesome")
}
