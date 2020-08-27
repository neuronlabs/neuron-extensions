/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/neuronlabs/neuron-extensions/internal/modules"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runTestSubmodules,
}

func init() {
	modulesCmd.AddCommand(testCmd)
}

func runTestSubmodules(_ *cobra.Command, args []string) {
	path, err := getExtensionsPath()
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}

	neuronPath, err := modulesCmd.PersistentFlags().GetString("neuron-path")
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
	if neuronPath == "" {
		neuronPath = filepath.Clean(path + "/../neuron")
	}

	var subModules []*modules.SubModule
	all, err := modulesCmd.PersistentFlags().GetBool("all")
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
	if all {
		subModules, err = modules.ListSubmodules(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		if len(args) == 0 {
			fmt.Printf("No modules provided to set on develop-mode. Provide submodules as an additional arguments or set the 'all' flag.")
			os.Exit(1)
		}
		for _, arg := range args {
			modulePath := arg
			moduleName := arg
			if strings.Contains(arg, extensionsPackage) {
				modulePath = strings.TrimPrefix(modulePath, extensionsPackage)
			} else {
				moduleName = extensionsPackage + "/" + moduleName
			}
			if modulePath[0] != '/' {
				modulePath = "/" + modulePath
			}
			modulePath = path + modulePath
			subModules = []*modules.SubModule{{Path: modulePath, ModuleName: moduleName}}
		}
	}
	for _, subModule := range subModules {
		if subModule.ModuleName == extensionsPackage {
			continue
		}
		if err := modules.TestModule(subModule); err != nil {
			fmt.Printf("Err: %v\n", err)
			os.Exit(1)
		}
	}
}
