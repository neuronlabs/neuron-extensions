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

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runUpdateModule,
}

func init() {
	modulesCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.
}

// modules update codec/json github.com/neuronlabs/neuron v1.0.0
func runUpdateModule(cmd *cobra.Command, args []string) {
	path, err := getExtensionsPath()
	if err != nil {
		os.Exit(1)
	}

	neuronPath, err := modulesCmd.PersistentFlags().GetString("neuron-path")
	if err != nil {
		fmt.Printf("Err: %v", err)
		os.Exit(1)
	}
	if neuronPath == "" {
		neuronPath = filepath.Clean(path + "/../neuron")
	}

	all, err := modulesCmd.PersistentFlags().GetBool("all")
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
	var (
		subModules                    []*modules.SubModule
		moduleToUpdate, moduleVersion string
	)
	if all {
		if len(args) < 2 {
			fmt.Printf("Not enough arguments provided. Provide module to update and it's version as arguments.")
			os.Exit(1)
		}
		subModules, err = modules.ListSubmodules(path)
		if err != nil {
			fmt.Printf("Err: %v\n", err)
			os.Exit(1)
		}
		moduleToUpdate = args[0]
		moduleVersion = args[1]
	} else {
		if len(args) < 3 {
			fmt.Printf("Not enough arguments provided. Provide submodule (or set the 'all' flag), module to update and it's version as arguments.")
			os.Exit(1)
		}
		modulePath := args[0]
		moduleName := args[0]
		if strings.Contains(modulePath, extensionsPackage) {
			modulePath = strings.TrimPrefix(modulePath, extensionsPackage)
		} else {
			moduleName = extensionsPackage + "/" + moduleName
		}
		if modulePath[0] != '/' {
			modulePath = "/" + modulePath
		}
		modulePath = path + modulePath
		moduleToUpdate = args[1]
		moduleVersion = args[2]
		subModules = []*modules.SubModule{{Path: modulePath, ModuleName: moduleName}}
	}
	for _, subModule := range subModules {
		if err = modules.UpdateModuleVersion(subModule, moduleToUpdate, moduleVersion); err != nil {
			fmt.Printf("Err: %v\n", err)
			os.Exit(1)
		}
	}
}
