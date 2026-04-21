/*
Copyright 2022 The Karmada Authors.

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
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	agentapp "github.com/karmada-io/karmada/cmd/agent/app"
	aaapp "github.com/karmada-io/karmada/cmd/aggregated-apiserver/app"
	cmapp "github.com/karmada-io/karmada/cmd/controller-manager/app"
	deschapp "github.com/karmada-io/karmada/cmd/descheduler/app"
	searchapp "github.com/karmada-io/karmada/cmd/karmada-search/app"
	adapterapp "github.com/karmada-io/karmada/cmd/metrics-adapter/app"
	estiapp "github.com/karmada-io/karmada/cmd/scheduler-estimator/app"
	schapp "github.com/karmada-io/karmada/cmd/scheduler/app"
	webhookapp "github.com/karmada-io/karmada/cmd/webhook/app"
	"github.com/karmada-io/karmada/pkg/karmadactl"
	"github.com/karmada-io/karmada/pkg/util/lifted"
	"github.com/karmada-io/karmada/pkg/util/names"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `usage: %s <output-directory> <component> [subcommand ...]

Generate cobra markdown flag/command reference for a Karmada component or CLI.

  <component>           One of the supported component names (see below).
  [subcommand ...]     Optional path from the component root command (e.g.
                       "create namespace" for karmadactl). When omitted, docs
                       are generated from the root command (full tree when the
                       root has subcommands).

Supported components:
`, os.Args[0])
	for _, name := range sortedComponentNames() {
		fmt.Fprintf(os.Stderr, "  - %s\n", name)
	}
}

func sortedComponentNames() []string {
	namesList := make([]string, 0, len(componentFactories))
	for n := range componentFactories {
		namesList = append(namesList, n)
	}
	sort.Strings(namesList)
	return namesList
}

// componentFactories maps the stable component name to a constructor for its root *cobra.Command.
var componentFactories = map[string]func() *cobra.Command{
	names.KarmadaControllerManagerComponentName: func() *cobra.Command {
		return cmapp.NewControllerManagerCommand(context.TODO())
	},
	names.KarmadaSchedulerComponentName: func() *cobra.Command {
		return schapp.NewSchedulerCommand(context.TODO())
	},
	names.KarmadaAgentComponentName: func() *cobra.Command {
		return agentapp.NewAgentCommand(context.TODO())
	},
	names.KarmadaAggregatedAPIServerComponentName: func() *cobra.Command {
		return aaapp.NewAggregatedApiserverCommand(context.TODO())
	},
	names.KarmadaDeschedulerComponentName: func() *cobra.Command {
		return deschapp.NewDeschedulerCommand(context.TODO())
	},
	names.KarmadaSearchComponentName: func() *cobra.Command {
		return searchapp.NewKarmadaSearchCommand(context.TODO())
	},
	names.KarmadaSchedulerEstimatorComponentName: func() *cobra.Command {
		return estiapp.NewSchedulerEstimatorCommand(context.TODO())
	},
	names.KarmadaWebhookComponentName: func() *cobra.Command {
		return webhookapp.NewWebhookCommand(context.TODO())
	},
	names.KarmadaMetricsAdapterComponentName: func() *cobra.Command {
		return adapterapp.NewMetricsAdapterCommand(context.TODO())
	},
	names.KarmadactlComponentName: func() *cobra.Command {
		return karmadactl.NewKarmadaCtlCommand(names.KarmadactlComponentName, names.KarmadactlComponentName)
	},
}

func findSubcommand(root *cobra.Command, parts []string) (*cobra.Command, error) {
	cur := root
	for _, want := range parts {
		var next *cobra.Command
		for _, child := range cur.Commands() {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			if child.Name() == want {
				next = child
				break
			}
			for _, a := range child.Aliases {
				if a == want {
					next = child
					break
				}
			}
			if next != nil {
				break
			}
		}
		if next == nil {
			return nil, fmt.Errorf("subcommand %q not found under %q", want, cur.CommandPath())
		}
		cur = next
	}
	return cur, nil
}

func main() {
	// use os.Args instead of "flags" because "flags" will mess up the man pages!
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	path := os.Args[1]
	component := os.Args[2]
	sub := os.Args[3:]

	outDir, err := lifted.OutDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get output directory: %v\n", err)
		os.Exit(1)
	}

	factory, ok := componentFactories[component]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown component %q (see usage)\n", component)
		printUsage()
		os.Exit(1)
	}

	// Construct the root command once: karmadactl registers klog flags at construction time and panics if built twice in one process.
	rootCmd := factory()
	workCmd := rootCmd
	if len(sub) > 0 {
		workCmd, err = findSubcommand(rootCmd, sub)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	workCmd.DisableAutoGenTag = true
	if err = doc.GenMarkdownTree(workCmd, outDir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate docs: %v\n", err)
		os.Exit(1)
	}

	if len(sub) == 0 {
		if err = tryGenGroupedCommandIndex(rootCmd, outDir); err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate grouped command index: %v\n", err)
			os.Exit(1)
		}
	}

	indexBasename := ""
	indexLinkText := ""
	if hasGroupedSubcommands(rootCmd) {
		indexBasename = strings.ReplaceAll(rootCmd.CommandPath(), " ", "_") + "_index.md"
		indexLinkText = indexPageTitle(rootCmd)
	}
	proc := func(md string) string {
		if indexBasename != "" && indexLinkText != "" {
			return cleanupForIncludeWithIndex(md, indexBasename, indexLinkText)
		}
		return cleanupForInclude(md)
	}
	if err = MarkdownPostProcessing(workCmd, outDir, proc); err != nil {
		fmt.Fprintf(os.Stderr, "failed to cleanup docs: %v\n", err)
		os.Exit(1)
	}
}
