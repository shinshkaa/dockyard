package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Application struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Port     int    `json:"port"`
	Replicas int    `json:"replicas"`
}

func (app Application) validateReplicas() bool {
	return app.Replicas > 0
}

func (app Application) printAppInfo() {
	if !app.validateReplicas() {
		fmt.Println("Invalid number of replicas")
		return
	}
	fmt.Println("Name:", app.Name)
	fmt.Println("Image:", app.Image)
	fmt.Println("Port:", app.Port)
	fmt.Println("Replicas:", app.Replicas)
}

func (app *Application) scale(replicas int) error {
	if replicas < 1 {
		return fmt.Errorf("invalid replicas count: %d", replicas)
	}
	app.Replicas = replicas
	return nil
}

func printAllApps(apps map[string]Application) {
	for _, app := range apps {
		app.printAppInfo()
		fmt.Println()
	}
}

func findApp(apps map[string]Application, key string) (Application, bool) {
	app, exists := apps[key]
	return app, exists
}

func removeApp(apps map[string]Application, name string) error {
	if _, exists := findApp(apps, name); exists {
		delete(apps, name)
		return nil
	}
	return errors.New("app doesn't exist")
}

func scaleApp(apps map[string]Application, name string, replicas int) error {
	if app, exists := apps[name]; exists {
		if err := app.scale(replicas); err != nil {
			return fmt.Errorf("scale error: %w", err)
		}
		apps[name] = app
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

func saveApps(apps map[string]Application) error {
	appsJson, err := json.MarshalIndent(apps, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile("apps.json", appsJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadApps() (map[string]Application, error) {
	appsJson, err := os.ReadFile("apps.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]Application), nil
		}
		return nil, err
	}
	apps := map[string]Application{}
	err = json.Unmarshal(appsJson, &apps)
	if err != nil {
		return nil, err
	}
	if apps == nil {
		return nil, errors.New("apps.json contains null instead of an application registry")
	}
	return apps, nil
}

func main() {
	apps, err := loadApps()
	if err != nil {
		fmt.Println("Loading error:", err)
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: dockyard <list|scale|remove>")
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
		if err = saveApps(apps); err != nil {
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
		if err = saveApps(apps); err != nil {
			fmt.Println(err)
			return
		}
	default:
		fmt.Println("Unknown command")
	}
}
