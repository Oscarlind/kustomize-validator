// Package validate provides content validation for rendered Kustomize manifests.
//
// The validateContent function searches for patterns in the rendered output and reports
// errors when matches are found. This is useful for catching placeholder values, debug
// settings, or other patterns that shouldn't appear in production.
//
// Pattern matching modes:
//   - Default (literal): Simple substring matching, e.g., "PATCH_ME"
//   - Glob: Wildcard matching, e.g., "glob:PATCH_*"
//   - Regex: Regular expression, e.g., "regex:\bPATCH_ME\b"
//
// Example usage:
//
//	carrier := Carrier{Path: "/path", Stdout: renderedManifest}
//	errors := ValidateContent(carrier, []string{"PATCH_ME", "glob:TODO*", "regex:\\blatest\\b"})
//	for _, err := range errors {
//	    fmt.Print(err.FormatError(true)) // verbose mode shows context
//	}
package validate

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Oscarlind/kustomize-validator/k8s"
	"github.com/gobwas/glob"
)

type Resource struct {
	k8s.Resource
	// Pattern the resource matched
	Pattern string
	// LineNumber the pattern was found on
	LineNumber int
	// MatchedLine the pattern was found in
	MatchedLine string
	// Context surrounding the matched line
	Context []string
}

type Resources []Resource

func (ve Resources) Error() error {
	var msgs []string
	for _, err := range ve {
		msgs = append(msgs, err.Error())
	}
	return errors.New(strings.Join(msgs, "\n"))
}

func (r Resources) Find(apiVersion, kind, namespace, name string) *Resource {
	for _, res := range r {
		if res.ApiVersion == apiVersion && res.Kind == kind && res.Namespace == namespace && res.Name == name {
			return &res
		}
	}
	return &Resource{}
}

// Error implements the error interface
func (e *Resource) Error() string {
	if e == nil || e.Pattern == "" {
		return ""
	}
	return fmt.Sprintf("validation failed: found '%s' in line %d for resource %s/%s/%s/%s", e.Pattern, e.LineNumber, e.ApiVersion, e.Kind, e.Namespace, e.Name)
}

// FormatError formats the validation error for display
func (e *Resource) FormatError(verbose bool) string {
	var output strings.Builder

	output.WriteString(Errorf("Content validation failed for apiVersion %s, kind %s, namespace %s, name %s", e.ApiVersion, e.Kind, e.Namespace, e.Name))
	output.WriteString(fmt.Sprintf("\tPattern: %s\n", e.Pattern))
	output.WriteString(fmt.Sprintf("\tLine: %d\n", e.LineNumber))
	output.WriteString(fmt.Sprintf("\tMatch: %s\n", strings.TrimSpace(e.MatchedLine)))

	// In verbose mode, show context
	if verbose && len(e.Context) > 0 {
		output.WriteString("\n\tContext:\n")
		startLine := e.LineNumber - (len(e.Context) / 2)
		if startLine < 1 {
			startLine = 1
		}

		for i, line := range e.Context {
			lineNum := startLine + i
			marker := "  "
			if lineNum == e.LineNumber {
				marker = "→ " // Highlight the matching line
			}
			output.WriteString(fmt.Sprintf("    %s%4d | %s\n", marker, lineNum, line))
		}
	}

	output.WriteString("\n")
	return output.String()
}

// ValidateContent validates the rendered kustomize output against multiple check patterns
func ValidateContent(resources []k8s.Resource, checks []string) Resources {
	var errors Resources
	for _, check := range checks {
		for _, resource := range resources {
			if err := validateContent(resource, check); err != nil {
				errors = append(errors, *err)
			}
		}
	}
	return errors
}

// validateContent is a helper function for individual content validation
// Returns a Resource if a match is found, nil otherwise
// Supports:
//   - literal substring matching (default)
//   - glob pattern matching (prefix with "glob:")
//   - regex pattern matching (prefix with "regex:")
func validateContent(resource k8s.Resource, check string) *Resource {
	var pattern string
	var matchFunc func(line string) bool

	switch {
	case strings.HasPrefix(check, "glob:"):
		pattern = strings.TrimPrefix(check, "glob:")
		matchFunc = createGlobMatcher(pattern)
	case strings.HasPrefix(check, "regex:"):
		pattern = strings.TrimPrefix(check, "regex:")
		matchFunc = createRegexMatcher(pattern)
	default:
		// Default: literal substring matching
		pattern = check
		matchFunc = createLiteralMatcher(pattern)
	}

	// Split output into lines for line-by-line matching
	lines := strings.Split(resource.FileContent, "\n")
	// Check each line
	for lineNum, line := range lines {
		if matchFunc(line) {
			// Extract context (±2 lines)
			context := extractContext(lines, lineNum, 2)
			return &Resource{
				Resource:    resource,
				Pattern:     pattern,
				LineNumber:  lineNum + 1,
				MatchedLine: line,
				Context:     context,
			}
		}
	}
	return nil
}

// createLiteralMatcher creates a matcher function for literal substring matching
func createLiteralMatcher(pattern string) func(string) bool {
	return func(line string) bool {
		return strings.Contains(line, pattern)
	}
}

// createGlobMatcher creates a matcher function for glob pattern matching
func createGlobMatcher(pattern string) func(string) bool {
	return func(line string) bool {
		return glob.MustCompile(pattern).Match(line)
	}
}

// createRegexMatcher creates a matcher function for regex pattern matching
func createRegexMatcher(pattern string) func(string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		// If regex compilation fails, fall back to literal matching
		return createLiteralMatcher(pattern)
	}

	return func(line string) bool {
		return re.MatchString(line)
	}
}

// extractContext extracts surrounding lines for context
func extractContext(lines []string, targetLine int, contextSize int) []string {
	start := targetLine - contextSize
	if start < 0 {
		start = 0
	}

	end := targetLine + contextSize + 1
	if end > len(lines) {
		end = len(lines)
	}

	context := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		context = append(context, lines[i])
	}

	return context
}
