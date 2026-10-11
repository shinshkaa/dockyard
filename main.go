package main

import (
	"dockyard/internal/app"
	"dockyard/internal/docker"
	"dockyard/internal/storage"
	"errors"
	"fmt"
	"os"
	"strconv"
)

func printAllApps(apps map[string]app.Application) {
	for _, localApp := range apps {
		localApp.PrintAppInfo()
		fmt.Println()
	}
}

func findApp(apps map[string]app.Application, key string) (app.Application, bool) {
	localApp, exists := apps[key]
	return localApp, exists
}

func removeApp(apps map[string]app.Application, name string) error {
	if _, exists := findApp(apps, name); exists {
		delete(apps, name)
		return nil
	}
	return errors.New("app doesn't exist")
}

func scaleApp(apps map[string]app.Application, name string, replicas int) error {
	if localApp, exists := apps[name]; exists {
		if err := localApp.Scale(replicas); err != nil {
			return fmt.Errorf("scale error: %w", err)
		}
		apps[name] = localApp
		return nil
	}
	return errors.New("app doesn't exist")
}

func checkLength(command []string, length int) error {
	if len(command) < length+1 {
		return errors.New("wrong command")
	}
	return nil
}

func main() {
	apps, err := storage.LoadApps()
	if err != nil {
		fmt.Println("Loading error:", err)
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: dockyard <list|scale|remove|containers|deploy|stop|delete>")
		return
	}
	switch os.Args[1] {
	case "list":
		printAllApps(apps)
	case "scale":
		if err := checkLength(os.Args, 3); err != nil {
			fmt.Printf("%v: scale <name> <replicas>\n", err)
			return
		}
		replicas, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Println("Invalid replicas count:", err)
			return
		}
		err = scaleApp(apps, os.Args[2], replicas)
		if err != nil {
			fmt.Println("Scaling error:", err)
			return
		}
		if err = storage.SaveApps(apps); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Application %s scaled to %d replicas\n\n", os.Args[2], replicas)
	case "remove":
		if err := checkLength(os.Args, 2); err != nil {
			fmt.Printf("%v: remove <name>\n", err)
			return
		}
		if err = removeApp(apps, os.Args[2]); err != nil {
			fmt.Println(err)
			return
		}
		if err = storage.SaveApps(apps); err != nil {
			fmt.Println(err)
			return
		}
	case "containers":
		err := docker.ListContainers()
		if err != nil {
			fmt.Println(err)
			return
		}
	case "deploy":
		if err := checkLength(os.Args, 4); err != nil {
			fmt.Printf("%v: deploy <image> <name> <port>\n", err)
			return
		}
		port, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Printf("Port is incorrect: %v\n", err)
			return
		}
		if port < 1 || port > 65535 {
			fmt.Println("Port must be between 1 and 65535")
			return
		}
		err = docker.DeployContainer(os.Args[2], os.Args[3], port)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Container %s deployed successfully\n", os.Args[3])
	case "stop":
		if err := checkLength(os.Args, 2); err != nil {
			fmt.Printf("%v: stop <name, id>\n", err)
			return
		}
		err = docker.StopContainer(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Container %s stopped successfully\n", os.Args[2])
	case "delete":
		if err := checkLength(os.Args, 2); err != nil {
			fmt.Printf("%v: delete <name, id>\n", err)
			return
		}
		err = docker.DeleteContainer(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Container %s deleted successfully\n", os.Args[2])
	default:
		fmt.Println("Unknown command")
	}
}
