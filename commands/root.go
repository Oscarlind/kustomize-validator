package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Oscarlind/kustomize-validator/validate"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var RootCmd = &cobra.Command{
	Use:  "kustomization-validator",
	Long: "A tool to validate Kustomization files",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No arguments provided")
			return
		}

		fmt.Println("Validating Kustomization files", args[0])
		isVerbose := cmd.Flag("verbose").Value.String() == "true"
		isErrorOnly := cmd.Flag("error-only").Value.String() == "true"
		isTable := cmd.Flag("table").Value.String() == "true"

		var tableRows [][]string
		cwd, _ := os.Getwd()

		if isTable {
			tableRows = append(tableRows, []string{"Relative path", "ApiVersion", "Kind", "Name", "Namespace"})
		}

		msgChan := validate.KustomizeBuild(args[0])
		ctx, cf := context.WithTimeout(context.Background(), 2*time.Second)
		defer cf()

		totalCounter := 0
		successCounter := 0
		failureCounter := 0

	BREAK:
		for {
			select {
			case <-ctx.Done():
				break BREAK
			case msg, ok := <-msgChan:
				if !ok {
					break BREAK
				}
				totalCounter++
				if msg.Err != nil {
					failureCounter++
				} else {
					successCounter++
				}

				if isTable {
					resources := parseKustomizeOutput(msg.Stdout, msg.Path, cwd)
					for _, resource := range resources {
						tableRows = append(tableRows, []string{
							resource.SourcePath,
							resource.ApiVersion,
							resource.Kind,
							resource.Name,
							resource.Namespace,
						})
					}
				} else {
					fmt.Print(msg.Msg(isErrorOnly, isVerbose))
				}
			}
		}

		if isTable && len(tableRows) > 1 {
			table := tablewriter.NewTable(os.Stdout)

			// Convert header to []any
			header := make([]any, len(tableRows[0]))
			for i, v := range tableRows[0] {
				header[i] = v
			}
			table.Header(header...)

			// Convert data rows to [][]any
			data := make([][]any, len(tableRows)-1)
			for i := 1; i < len(tableRows); i++ {
				row := make([]any, len(tableRows[i]))
				for j, v := range tableRows[i] {
					row[j] = v
				}
				data[i-1] = row
			}
			table.Bulk(data)

			table.Render()
		}

		if !isTable {
			fmt.Println("Total: ", validate.ColorF(validate.ColorBlue, "%d", totalCounter))
			fmt.Println("Success: ", validate.ColorF(validate.ColorGreen, "%d", successCounter))
			fmt.Println("Error: ", validate.ColorF(validate.ColorRed, "%d", failureCounter))
			fmt.Println("Failed in %: ", validate.ColorF(validate.ColorRed, "%.2f%%", float64(failureCounter)/float64(totalCounter)*100))
		}
	},
}

type Resource struct {
	ApiVersion string
	Kind       string
	Name       string
	Namespace  string
	SourcePath string
}

type KubernetesResource struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
}

func parseKustomizeOutput(stdout, sourcePath, cwd string) []Resource {
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
			namespace = "default"
		}

		relativePath, err := filepath.Rel(cwd, sourcePath)
		if err != nil {
			relativePath = sourcePath
		}

		resources = append(resources, Resource{
			ApiVersion: resource.ApiVersion,
			Kind:       resource.Kind,
			Name:       resource.Metadata.Name,
			Namespace:  namespace,
			SourcePath: relativePath,
		})
	}

	return resources
}

func shortenPath(fullPath string) string {
	path := strings.TrimPrefix(fullPath, "overlays/")
	parts := strings.Split(path, "/")

	if len(parts) > 4 {
		return fmt.Sprintf("%s/%s/.../%s/%s", parts[0], parts[1], parts[len(parts)-2], parts[len(parts)-1])
	}
	if len(parts) > 2 {
		return path
	}
	return path
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolP("error-only", "e", false, "whether we should only log errors")
	RootCmd.PersistentFlags().BoolP("table", "t", false, "output resources in table format")
}
