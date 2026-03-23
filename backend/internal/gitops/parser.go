package gitops

import (
	"bytes"
	"io"

	"github.com/hemu1808/auradeploy/backend/internal/models"
	"gopkg.in/yaml.v3"
)

type Manifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		Image    string `yaml:"image"`
		Replicas int    `yaml:"replicas"`
		Env      []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"env"`
		Ports []struct {
			ContainerPort int `yaml:"containerPort"`
			HostPort      int `yaml:"hostPort"`
		} `yaml:"ports"`
	} `yaml:"spec"`
}

// ParseManifests decodes a multi-document YAML into internal Application models
func ParseManifests(data []byte) ([]models.Application, error) {
	var apps []models.Application
	dec := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var m Manifest
		err := dec.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if m.Kind != "Application" || m.APIVersion != "aura.deploy/v1" {
			continue // skip unknown kinds
		}

		var envs []models.EnvVar
		for _, e := range m.Spec.Env {
			envs = append(envs, models.EnvVar{Key: e.Name, Value: e.Value})
		}

		var ports []models.PortMapping
		for _, p := range m.Spec.Ports {
			ports = append(ports, models.PortMapping{HostPort: p.HostPort, ContainerPort: p.ContainerPort})
		}

		app := models.Application{
			ID:          m.Metadata.Name,
			Name:        m.Metadata.Name,
			DockerImage: m.Spec.Image,
			Replicas:    models.ReplicasJSON{Target: m.Spec.Replicas},
			EnvVars:     envs,
			Ports:       ports,
		}
		apps = append(apps, app)
	}

	return apps, nil
}
