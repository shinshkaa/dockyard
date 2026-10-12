package app

import "fmt"

type Application struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Port     int    `json:"port"`
	Replicas int    `json:"replicas"`
	Network  string `json:"network"`
}

func (app *Application) Scale(replicas int) error {
	if replicas < 1 {
		return fmt.Errorf("invalid replicas count: %d", replicas)
	}
	app.Replicas = replicas
	return nil
}

func (app Application) ValidateReplicas() bool {
	return app.Replicas > 0
}

func (app Application) PrintAppInfo() {
	fmt.Println("Name:", app.Name)
	fmt.Println("Image:", app.Image)
	fmt.Println("Port:", app.Port)
	fmt.Println("Replicas:", app.Replicas)
}
