package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const (
	Green       = "\033[0;32m"
	Red         = "\033[0;31m"
	Yellow      = "\033[0;33m"
	NC          = "\033[0m" // No Color
	FailedColor = Yellow   // Color for highlighting the failed path
)

func main() {
	var wg sync.WaitGroup
	okCount := 0
	failedCount := 0
	hasFailure := false // Flag to track if any failures occurred

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}
		if info.Name() == "kustomization.yaml" {
			dir := filepath.Dir(path)
			wg.Add(1)
			go func(dir string) {
				defer wg.Done()
				cmd := exec.Command("kustomize", "build", "--enable-helm", "--enable-alpha-plugins", dir) // allows for Helm and policyGenerator plugin
				output, err := cmd.CombinedOutput()
				if err == nil {
					fmt.Printf("%sOK%s: %s\n", Green, NC, dir)
					okCount++
				} else {
					fmt.Printf("%sFAILED%s: %s\n", Red, NC, dir)
					failedCount++
					hasFailure = true // Set the failure flag
					failedPaths := extractFailedPaths(string(output))
					for _, path := range failedPaths {
						fullPath := filepath.Join(dir, path)
						fmt.Printf("Failed path: %s%s%s\n", FailedColor, fullPath, NC)
					}
				}
			}(dir)
		}
		return nil
	})

	wg.Wait()

	separator := strings.Repeat("=", 20)
	fmt.Println(separator)
	fmt.Printf("Total OK: %d\n", okCount)
	fmt.Printf("Total FAILED: %d\n", failedCount)

	if hasFailure {
		os.Exit(1) // Exit with non-zero return code
	}
}

// extractFailedPaths extracts the file paths mentioned in the error message.
func extractFailedPaths(output string) []string {
	regex := regexp.MustCompile(`from '(.*?)'`)
	matches := regex.FindAllStringSubmatch(output, -1)
	var paths []string
	for _, match := range matches {
		if len(match) > 1 {
			paths = append(paths, match[1])
		}
	}
	return paths
}