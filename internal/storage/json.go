package storage

import (
	"dockyard/internal/app"
	"encoding/json"
	"errors"
	"os"
)

func SaveApps(apps map[string]app.Application) error {
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

func LoadApps() (map[string]app.Application, error) {
	appsJson, err := os.ReadFile("apps.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]app.Application), nil
		}
		return nil, err
	}
	apps := map[string]app.Application{}
	err = json.Unmarshal(appsJson, &apps)
	if err != nil {
		return nil, err
	}
	if apps == nil {
		return nil, errors.New("apps.json contains null instead of an application registry")
	}
	return apps, nil
}
