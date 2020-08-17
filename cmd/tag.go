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
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// tagCmd represents the tag command
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Creates a tag version for the submodule",
	Long: `In order to set up golang submodule version, a git repository needs to have a tag 
that is composed of submodule name and tag version, i.e.: codec/json/v0.0.1.
This command allows to create it easily by providing two arguments: 
- module name
- new version
- optional message`,
	Run: runTagCommand,
}

func runTagCommand(_ *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println("provided to few arguments.")
		os.Exit(1)
	}
	moduleName := args[0]
	moduleName = strings.TrimPrefix(moduleName, "github.com/neuronlabs/neuron-extensions")
	moduleVersion := args[1]
	if moduleVersion[0] != 'v' {
		fmt.Printf("provided invalid semantic version: %v\n", moduleVersion)
		os.Exit(1)
	}
	var message string
	if len(args) >= 3 {
		message = "\"" + strings.Join(args[2:], " ") + "\""
	}

	tagName := moduleName + "/" + moduleVersion
	gitArgs := []string{"tag"}
	if len(args) > 3 {
		gitArgs = append(gitArgs, "-a")
	}
	gitArgs = append(gitArgs, tagName)
	if len(args) > 3 {
		gitArgs = append(gitArgs, "-m", message)
	}

	cmd := exec.Command("git", gitArgs...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Git failed: %v\n", err)
		os.Exit(1)
	}
	if len(out) > 0 {
		fmt.Println(out)
	}
}

func init() {
	modulesCmd.AddCommand(tagCmd)
}
