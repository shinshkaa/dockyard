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
		currentApp, exists := apps[os.Args[2]]
		if !exists {
			fmt.Println("Application doesn't exist")
			return
		}

		if replicas < 1 {
			fmt.Println("Replicas must be greater than zero")
			return
		} else if replicas < currentApp.Replicas {
			if err = docker.ScaleDown(
				currentApp.Name, currentApp.Replicas, replicas,
			); err != nil {
				fmt.Println(err)
				return
			}
		} else {
			if err = docker.ScaleUp(
				currentApp.Image, currentApp.Name, currentApp.Network, currentApp.Replicas, replicas,
			); err != nil {
				fmt.Println(err)
				return
			}
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
		fmt.Printf("Application %s scaled to %d replicas\n", os.Args[2], replicas)
	case "containers":
		err := docker.ListContainers()
		if err != nil {
			fmt.Println(err)
			return
		}
	case "deploy":
		if err := checkLength(os.Args, 5); err != nil {
			fmt.Printf("%v: deploy <image> <name> <port> <network_name>\n", err)
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
		newApp := app.Application{
			Name:     os.Args[3],
			Image:    os.Args[2],
			Port:     port,
			Replicas: 1,
			Network:  os.Args[5],
		}
		if _, exists := apps[os.Args[3]]; exists {
			fmt.Println("Application already exists")
			return
		}

		containerName := fmt.Sprintf("%s-1", newApp.Name)

		err = docker.DeployContainer(newApp.Image, containerName, newApp.Network, newApp.Port)
		if err != nil {
			fmt.Println(err)
			return
		}
		apps[newApp.Name] = newApp
		if err := storage.SaveApps(apps); err != nil {
			fmt.Println("Saving error:", err)
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
			fmt.Printf("%v: delete <name>\n", err)
			return
		}

		currentApp, exists := findApp(apps, os.Args[2])
		if !exists {
			fmt.Println("App doesn't exist")
			return
		}

		err := docker.DeleteApplication(currentApp.Name, currentApp.Replicas)
		if err != nil {
			fmt.Println(err)
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
		fmt.Println("App deleted successfully")
	case "logs":
		if err := checkLength(os.Args, 2); err != nil {
			fmt.Printf("%v: logs <name, id>\n", err)
			return
		}
		if err = docker.ShowLogs(os.Args[2]); err != nil {
			fmt.Printf("Logs reading error: %v\n", err)
			return
		}
	case "network":
		if err := checkLength(os.Args, 2); err != nil {
			fmt.Printf("%v: network <create>\n", err)
			return
		}
		switch os.Args[2] {
		case "create":
			if err := checkLength(os.Args, 3); err != nil {
				fmt.Printf("%v: network create <name>\n", err)
				return
			}
			if err = docker.CreateNetwork(os.Args[3]); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Network %s successfully created\n", os.Args[3])
		case "connect":
			if err := checkLength(os.Args, 4); err != nil {
				fmt.Printf("%v: network connect <network_name> <container_name>\n", err)
				return
			}
			if err = docker.ConnectNetwork(os.Args[3], os.Args[4]); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Container %s successfully connected to %s network\n", os.Args[4], os.Args[3])
		}
	case "replicas":
		if err := checkLength(os.Args, 5); err != nil {
			fmt.Printf("%v: replicas <image> <name> <count> <network>\n", err)
			return
		}

		if _, exists := findApp(apps, os.Args[3]); exists {
			fmt.Println("Application already exists")
			return
		}

		replicasCount, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Printf("Wrong replicas count %s: %v\n", os.Args[4], err)
			return
		}
		if replicasCount < 1 {
			fmt.Println("Replicas count must be greater than zero")
			return
		}
		if err = docker.DeployReplicas(os.Args[2], os.Args[3], os.Args[5], replicasCount); err != nil {
			fmt.Println(err)
			return
		}

		newApp := app.Application{
			Name:     os.Args[3],
			Image:    os.Args[2],
			Port:     0,
			Replicas: replicasCount,
			Network:  os.Args[5],
		}
		apps[newApp.Name] = newApp
		if err := storage.SaveApps(apps); err != nil {
			fmt.Println("Saving error:", err)
			return
		}
		fmt.Printf("Container %s deployed successfully\n", os.Args[3])

		fmt.Printf("%d replicas created\n", replicasCount)
	default:
		fmt.Println("Unknown command")
	}
}
