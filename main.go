package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Application struct {
	Name     string
	Image    string
	Port     int
	Replicas int
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

func removeApp(apps map[string]Application, name string) {
	delete(apps, name)
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

func main() {
	apps := map[string]Application{
		"prometheus": {
			Name:     "Prometheus",
			Image:    "prom/prometheus",
			Port:     9090,
			Replicas: 3,
		},
		"grafana": {
			Name:     "Grafana",
			Image:    "grafana/grafana",
			Port:     3000,
			Replicas: 1,
		},
	}
	apps["nginx"] = Application{
		Name:     "Nginx",
		Image:    "nginx:alpine",
		Port:     8080,
		Replicas: 3,
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
		fmt.Printf("Application %s scaled to %d replicas\n\n", os.Args[2], replicas)
	case "remove":
		if err := checkLength(os.Args, 2); err != nil {
			fmt.Printf("%v: remove <name>\n", err)
			return
		}
		removeApp(apps, os.Args[2])
	default:
		fmt.Println("Unknown command")
	}

	//app := apps["prometheus"]
	//
	//err := app.scale(0)
	//if err != nil {
	//	err = fmt.Errorf("cannot scale %s: %w", app.Name, err)
	//	fmt.Println("Error:", err)
	//	return
	//}
	//removeApp(apps, "nginx")
	//if app, isFound := findApp(apps, "nginx"); isFound {
	//	app.printAppInfo()
	//} else {
	//	fmt.Println("The app wasn't found")
	//}

	//err := scaleApp(apps, "grafana", 1)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//err = scaleApp(apps, "promtail", 3)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//err = scaleApp(apps, "grafana", -2)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
}
