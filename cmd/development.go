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

// developModeCmd represents the developMode command
var developModeCmd = &cobra.Command{
	Use:   "development",
	Short: "Replaces submodules go.mod files with the 'replace' clause.",
	Long: `This function iterates over all submodules recursively and finds all submodules. 
For each submodule it creates a copy of it's go.mod file, and creates a new one, 
where all neuron references are replaces using 'replace' clause. 
To undo this mode - execute production-mode. 
The function could set the production-mode for all submodules (using -all flag) or 
for the selected submodule by providing it as the argument.`,
	Run: runDevelopMode,
}

func init() {
	modulesCmd.AddCommand(developModeCmd)
}

// neuron-extensions modules develop-mode . codec/json
func runDevelopMode(_ *cobra.Command, args []string) {
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
			fmt.Printf("No modules provided to set on develop-mode. Provide submodules as arguments or set the 'all' flag.")
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
		if err = modules.SetDevelopmentMode(subModule, neuronPath, path); err != nil {
			if err == modules.ErrModuleAlreadyDevelopment {
				continue
			}
			fmt.Printf("Replacing development module failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func getExtensionsPath() (string, error) {
	path, err := modulesCmd.PersistentFlags().GetString("extensions-path")
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		return "", err
	}
	if path == "" {
		path = "."
	}
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Can't open input path: %v\n", err)
		return "", err
	}
	defer f.Close()
	dir, err := f.Stat()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if !dir.IsDir() {
		fmt.Println("provided input is not a directory")
		return "", fmt.Errorf("provided input is not a directory")
	}

	path, err = filepath.Abs(path)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if !strings.HasSuffix(path, "neuron-extensions") {
		err := fmt.Errorf("provided invalid input directory: %s\n", path)
		fmt.Println(err)
		return "", err
	}
	return path, nil
}
