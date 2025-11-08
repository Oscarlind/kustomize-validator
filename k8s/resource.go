package k8s

import (
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type KubernetesResource struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
}

type Resource struct {
	ApiVersion  string
	Kind        string
	Name        string
	Namespace   string
	SourcePath  string
	FileContent string
}

// ParseKustomizeOutput parses the kustomize output and returns a list of Resources
func ParseKustomizeOutput(stdout, sourcePath, cwd string) []Resource {
	if stdout == "" {
		return []Resource{}
	}

	var resources []Resource
	documents := strings.Split(stdout, "---")

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var resource KubernetesResource
		err := yaml.Unmarshal([]byte(doc), &resource)
		if err != nil {
			continue
		}

		if resource.ApiVersion == "" || resource.Kind == "" || resource.Metadata.Name == "" {
			continue
		}

		namespace := resource.Metadata.Namespace
		if namespace == "" {
			namespace = "<none>"
		}

		relativePath, err := filepath.Rel(cwd, sourcePath)
		if err != nil {
			relativePath = sourcePath
		}

		resources = append(resources, Resource{
			ApiVersion:  resource.ApiVersion,
			Kind:        resource.Kind,
			Name:        resource.Metadata.Name,
			Namespace:   namespace,
			SourcePath:  relativePath,
			FileContent: doc,
		})
	}

	return resources
}
