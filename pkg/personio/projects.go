package personio

import (
	"fmt"
	"net/http"
)

type Project struct {
	ID         int `json:"id"`
	Attributes struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	} `json:"attributes"`
}

func (client *Client) GetProjectID(name string) (int, error) {
	err := client.cacheProjects()
	if err != nil {
		return 0, err
	}
	for _, project := range client.projectCache {
		if project.Attributes.Name == name {
			return project.ID, nil
		}
	}
	return 0, fmt.Errorf("project %s not found", name)
}

func (client *Client) GetProjectName(id int) (string, error) {
	err := client.cacheProjects()
	if err != nil {
		return "", err
	}
	for _, project := range client.projectCache {
		if project.ID == id {
			return project.Attributes.Name, nil
		}
	}
	return "", fmt.Errorf("project %d not found", id)
}

func (client *Client) cacheProjects() error {
	if client.projectCache != nil {
		return nil
	}
	request, err := http.NewRequest("GET", "/api/v1/projects", nil)
	if err != nil {
		return err
	}
	resp, err := client.RawJSON(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	projects, err := ParseResponseJSON[[]Project](resp)
	if err != nil {
		return err
	}
	client.projectCache = projects
	return nil
}
