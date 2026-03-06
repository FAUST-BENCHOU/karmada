/*
Copyright 2024 The Karmada Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	controllerruntime "sigs.k8s.io/controller-runtime"

	controllermanagerapp "github.com/karmada-io/karmada/cmd/controller-manager/app"
)

type component struct {
	name    string
	command func(context.Context) *cobra.Command
}

var components = []component{
	{
		name:    "karmada-controller-manager",
		command: controllermanagerapp.NewControllerManagerCommand,
	},
}

// extractDeprecatedFlags extracts deprecated flags from a command
func extractDeprecatedFlags(cmd *cobra.Command) map[string]string {
	deprecated := make(map[string]string)
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Deprecated != "" {
			deprecated[flag.Name] = flag.Deprecated
		}
	})
	// Also check subcommands
	for _, subCmd := range cmd.Commands() {
		subDeprecated := extractDeprecatedFlags(subCmd)
		maps.Copy(deprecated, subDeprecated)
	}
	return deprecated
}

// formatDeprecatedFlags formats deprecated flags information
func formatDeprecatedFlags(deprecated map[string]string) string {
	if len(deprecated) == 0 {
		return ""
	}
	var lines []string
	lines = append(lines, "")
	lines = append(lines, "Deprecated flags:")
	lines = append(lines, "")
	for flagName, deprecationMsg := range deprecated {
		lines = append(lines, fmt.Sprintf("      [DEPRECATED] --%s", flagName))
		lines = append(lines, fmt.Sprintf("                           %s", deprecationMsg))
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func extractFlagsFromHelp(helpOutput string) string {
	lines := strings.Split(helpOutput, "\n")
	var flagLines []string
	inFlagsSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the beginning
		if !inFlagsSection && trimmed == "" {
			continue
		}

		// Start capturing when we see a flag section header
		if strings.HasSuffix(trimmed, "flags:") && !strings.Contains(trimmed, "Usage:") {
			inFlagsSection = true
			flagLines = append(flagLines, line)
			continue
		}

		// Stop if we hit subcommands or help topics
		if inFlagsSection {
			if strings.HasPrefix(trimmed, "Available Commands:") ||
				strings.HasPrefix(trimmed, "Use \"") ||
				strings.HasPrefix(trimmed, "Additional help topics:") {
				break
			}
			flagLines = append(flagLines, line)
		}
	}

	return strings.Join(flagLines, "\n")
}

func main() {
	ctx := controllerruntime.SetupSignalHandler()
	var output strings.Builder

	output.WriteString("Usage:\n   [flags]\n\n")

	for i, comp := range components {
		if i > 0 {
			output.WriteString("\n")
		}

		var buf bytes.Buffer
		cmd := comp.command(ctx)
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		// Call Help() to get formatted help output
		if err := cmd.Help(); err != nil {
			fmt.Fprintf(os.Stderr, "Error getting help for %s: %v\n", comp.name, err)
			continue
		}

		helpOutput := buf.String()
		flagsSection := extractFlagsFromHelp(helpOutput)

		// Extract deprecated flags
		deprecatedFlags := extractDeprecatedFlags(cmd)
		deprecatedSection := formatDeprecatedFlags(deprecatedFlags)

		if flagsSection != "" {
			// Add component name as header
			output.WriteString(fmt.Sprintf("%s flags:\n\n", comp.name))
			output.WriteString(flagsSection)

			// Add deprecated flags section if any
			if deprecatedSection != "" {
				output.WriteString("\n")
				output.WriteString(deprecatedSection)
			}

			output.WriteString("\n")
		} else {
			// Fallback: try to extract from the entire help output
			fmt.Fprintf(os.Stderr, "Warning: Could not extract flags section for %s, using full help output\n", comp.name)
			output.WriteString(fmt.Sprintf("%s flags:\n\n", comp.name))
			output.WriteString(helpOutput)

			// Add deprecated flags section if any
			if deprecatedSection != "" {
				output.WriteString("\n")
				output.WriteString(deprecatedSection)
			}

			output.WriteString("\n")
		}
	}

	// Write to stdout
	fmt.Print(output.String())
}
