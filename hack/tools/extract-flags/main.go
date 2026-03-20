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
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	_ "k8s.io/component-base/logs/json/register" // register JSON log format so --log-json-* flags match component binaries
	controllerruntime "sigs.k8s.io/controller-runtime"

	aggregatedapiserverapp "github.com/karmada-io/karmada/cmd/aggregated-apiserver/app"
	controllermanagerapp "github.com/karmada-io/karmada/cmd/controller-manager/app"
	deschedulerapp "github.com/karmada-io/karmada/cmd/descheduler/app"
	karmadasearchapp "github.com/karmada-io/karmada/cmd/karmada-search/app"
	metricsadapterapp "github.com/karmada-io/karmada/cmd/metrics-adapter/app"
	schedulerestimatorapp "github.com/karmada-io/karmada/cmd/scheduler-estimator/app"
	schedulerapp "github.com/karmada-io/karmada/cmd/scheduler/app"
	webhookapp "github.com/karmada-io/karmada/cmd/webhook/app"
	"github.com/karmada-io/karmada/pkg/karmadactl"
)

type component struct {
	name               string
	command            func(context.Context) *cobra.Command
	includeSubcommands bool
}

var components = []component{
	{
		name:    "karmada-controller-manager",
		command: controllermanagerapp.NewControllerManagerCommand,
	},
	{
		name: "karmada-scheduler",
		command: func(ctx context.Context) *cobra.Command {
			return schedulerapp.NewSchedulerCommand(ctx)
		},
	},
	{
		name: "karmada-search",
		command: func(ctx context.Context) *cobra.Command {
			return karmadasearchapp.NewKarmadaSearchCommand(ctx)
		},
	},
	{
		name:    "karmada-webhook",
		command: webhookapp.NewWebhookCommand,
	},
	{
		name:    "karmada-aggregated-apiserver",
		command: aggregatedapiserverapp.NewAggregatedApiserverCommand,
	},
	{
		name:    "karmada-descheduler",
		command: deschedulerapp.NewDeschedulerCommand,
	},
	{
		name:    "karmada-metrics-adapter",
		command: metricsadapterapp.NewMetricsAdapterCommand,
	},
	{
		name:    "karmada-scheduler-estimator",
		command: schedulerestimatorapp.NewSchedulerEstimatorCommand,
	},
	{
		name: "karmadactl",
		command: func(context.Context) *cobra.Command {
			return karmadactl.NewKarmadaCtlCommand("karmadactl", "karmadactl")
		},
		includeSubcommands: true,
	},
}

// formatDeprecatedFlags extracts and formats deprecated flags from a command and its subcommands.
func formatDeprecatedFlags(cmd *cobra.Command) string {
	var collect func(*cobra.Command) map[string]string
	collect = func(c *cobra.Command) map[string]string {
		deprecated := make(map[string]string)
		c.Flags().VisitAll(func(flag *pflag.Flag) {
			if flag.Deprecated != "" {
				deprecated[flag.Name] = flag.Deprecated
			}
		})
		for _, sub := range c.Commands() {
			maps.Copy(deprecated, collect(sub))
		}
		return deprecated
	}
	deprecated := collect(cmd)
	if len(deprecated) == 0 {
		return ""
	}
	names := make([]string, 0, len(deprecated))
	for name := range deprecated {
		names = append(names, name)
	}
	sort.Strings(names)
	var lines []string
	lines = append(lines, "", "Deprecated flags:", "")
	for _, flagName := range names {
		lines = append(lines, fmt.Sprintf("      [DEPRECATED] --%s", flagName), fmt.Sprintf("                           %s", deprecated[flagName]), "")
	}
	return strings.Join(lines, "\n")
}

func collectSubcommandHelp(cmd *cobra.Command, path string, out *strings.Builder) {
	for _, sub := range cmd.Commands() {
		if sub.Hidden {
			continue
		}
		subPath := path + " " + sub.Name()
		var buf bytes.Buffer
		sub.SetOut(&buf)
		sub.SetErr(&buf)
		if err := sub.Help(); err != nil {
			continue
		}
		out.WriteString("\n\n")
		out.WriteString("=== " + subPath + " ===\n\n")
		out.Write(buf.Bytes())
		if s := formatDeprecatedFlags(sub); s != "" {
			out.WriteString(s)
		}
		collectSubcommandHelp(sub, subPath, out)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: extract-flags <output-dir>")
		fmt.Fprintln(os.Stderr, "  e.g. extract-flags docs/command-flags")
		os.Exit(1)
	}
	outputDir := os.Args[1]

	ctx := controllerruntime.SetupSignalHandler()

	for _, comp := range components {
		cmd := comp.command(ctx)
		var content []byte

		cmd.InitDefaultHelpCmd()
		cmd.InitDefaultCompletionCmd()
		cmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		if err := cmd.Help(); err != nil {
			fmt.Fprintf(os.Stderr, "Error getting help for %s: %v\n", comp.name, err)
			continue
		}
		content = buf.Bytes()
		if s := formatDeprecatedFlags(cmd); s != "" {
			content = append(content, []byte(s)...)
		}
		if comp.includeSubcommands {
			var out strings.Builder
			out.Write(content)
			collectSubcommandHelp(cmd, comp.name, &out)
			content = []byte(out.String())
		}

		outputPath := filepath.Join(outputDir, comp.name+".txt")
		if err := os.WriteFile(outputPath, content, 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outputPath, err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %s\n", outputPath)
	}
}
