package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Oscarlind/kustomize-validator/validate"
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
		// Get current working directory for relative path calculation
		cwd, _ := os.Getwd()

		if isTable {
			// Add header
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
					// Parse resources and collect table rows
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

		// Print aligned table if in table mode
		if isTable {
			printAlignedTable(tableRows)
		}

		// Only show summary if not in table mode
		if !isTable {
			fmt.Println("Total: ", validate.ColorF(validate.ColorBlue, "%d", totalCounter))
			fmt.Println("Success: ", validate.ColorF(validate.ColorGreen, "%d", successCounter))
			fmt.Println("Error: ", validate.ColorF(validate.ColorRed, "%d", failureCounter))
			fmt.Println("Failed in %: ", validate.ColorF(validate.ColorRed, "%.2f%%", float64(failureCounter)/float64(totalCounter)*100))
		}
	},
}

// Resource represents a Kubernetes resource
type Resource struct {
	ApiVersion string
	Kind       string
	Name       string
	Namespace  string
	SourcePath string
}

// KubernetesResource represents the structure we expect from YAML
type KubernetesResource struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
}

// parseKustomizeOutput parses the YAML output from kustomize and extracts resource information
func parseKustomizeOutput(stdout, sourcePath, cwd string) []Resource {
	if stdout == "" {
		return []Resource{}
	}

	var resources []Resource
	
	// Split YAML documents (separated by ---)
	documents := strings.Split(stdout, "---")
	
	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var resource KubernetesResource
		err := yaml.Unmarshal([]byte(doc), &resource)
		if err != nil {
			// Skip invalid YAML documents
			continue
		}

		// Skip if essential fields are missing
		if resource.ApiVersion == "" || resource.Kind == "" || resource.Metadata.Name == "" {
			continue
		}

		// Set default namespace if empty
		namespace := resource.Metadata.Namespace
		if namespace == "" {
			namespace = "default"
		}

		// Calculate relative path
		relativePath, err := filepath.Rel(cwd, sourcePath)
		if err != nil {
			// If we can't calculate relative path, use the original
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

// printAlignedTable prints a properly aligned table
func printAlignedTable(rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(rows[0]))
	for _, row := range rows {
		for i, col := range row {
			if len(col) > colWidths[i] {
				colWidths[i] = len(col)
			}
		}
	}

	// Print header
	printTableRow(rows[0], colWidths, true)
	
	// Print separator
	printSeparator(colWidths)
	
	// Print data rows
	for i := 1; i < len(rows); i++ {
		printTableRow(rows[i], colWidths, false)
	}
}

// printTableRow prints a single table row with proper alignment
func printTableRow(row []string, widths []int, isHeader bool) {
	fmt.Print("| ")
	for i, col := range row {
		if isHeader {
			fmt.Printf("%-*s", widths[i], col)
		} else {
			fmt.Printf("%-*s", widths[i], col)
		}
		fmt.Print(" | ")
	}
	fmt.Println()
}

// printSeparator prints the table separator line
func printSeparator(widths []int) {
	fmt.Print("|")
	for _, width := range widths {
		fmt.Print(strings.Repeat("-", width+2) + "|")
	}
	fmt.Println()
}
func shortenPath(fullPath string) string {
	// Remove common prefixes
	path := strings.TrimPrefix(fullPath, "overlays/")
	
	// Split path into parts
	parts := strings.Split(path, "/")
	
	// If path is still too long, show only the most relevant parts
	if len(parts) > 4 {
		// Show first 2 and last 2 parts with ... in between
		return fmt.Sprintf("%s/%s/.../%s/%s", parts[0], parts[1], parts[len(parts)-2], parts[len(parts)-1])
	}
	
	// If path has 3-4 parts, show all
	if len(parts) > 2 {
		return path
	}
	
	// For very short paths, return as-is
	return path
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolP("error-only", "e", false, "whether we should only log errors")
	RootCmd.PersistentFlags().BoolP("table", "t", false, "output resources in table format")
}