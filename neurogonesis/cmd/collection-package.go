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
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neuronlabs/strcase"
	"github.com/spf13/cobra"

	"github.com/neuronlabs/neuron-extensions/neurogonesis/internal/ast"
)

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "creates an external package collection",
	Long: `This command creates an external package that contains collection for provided type
creates a private variable for given collection and exposes it's methods as the package public functions.`,

	// neurogonesis collections package User userdb
	Run:     collectionPackage,
	Example: "neurogonesis collections package User userdb",
}

func init() {
	collectionsCmd.AddCommand(packageCmd)

	// Here you will define your flags and configuration settings.
	packageCmd.Flags().StringP("input", "i", ".", "provide input directory")
}

func collectionPackage(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Printf("Collection package requires exactly 2 arguments")
		cmd.Usage()
		return
	}
	typeName := args[0]
	outputPackageName := args[1]
	tags := args[2:]
	namingConvention, err := collectionsCmd.PersistentFlags().GetString("naming-convention")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cmd.Usage()
		os.Exit(2)
	}

	input, err := cmd.Flags().GetString("input")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cmd.Usage()
		os.Exit(1)
	}
	if input == "" {
		input = "."
	}
	output, err := collectionsCmd.PersistentFlags().GetString("output")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cmd.Usage()
		os.Exit(1)
	}
	if output == "" {
		output = filepath.Join(".", outputPackageName)
	}
	output = filepath.Clean(output)
	if err := os.MkdirAll(output, 0777); err != nil && err != os.ErrExist {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}

	switch namingConvention {
	case "kebab", "snake", "lower_camel", "camel":
	default:
		fmt.Fprintf(os.Stderr, "Error: provided unsupported naming convention: '%v'", namingConvention)
		cmd.Usage()
		os.Exit(2)
	}
	g := ast.NewModelGenerator(namingConvention, []string{typeName}, tags, []string{})

	// Parse provided argument packages.
	g.ParsePackages(input)

	// Extract all models from given packages.
	if err := g.ExtractPackages(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Extract the directory name from the arguments.
	dir := output

	// Generate collection files.
	buf := &bytes.Buffer{}

	var (
		packageName string
	)
	v, _ := filepath.Rel(filepath.Clean("."), dir)
	isModelImported := v != "."
	if isModelImported {
		g.ResolveRelationSelectors()
		packageName = filepath.Base(filepath.Clean(dir))
	}

	pkgCollection, err := g.CollectionInput(packageName, isModelImported, typeName)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
	collectionPackageFileName := filepath.Join(dir, strcase.ToSnake(pkgCollection.Model.Name)+"_db.gen")
	collectionPackageFileName += ".go"
	generateFile(collectionPackageFileName, "collection-package", buf, pkgCollection)

	fmt.Fprintf(os.Stdout, "Success. Generated collection package: '%s' for: %s model.\n", outputPackageName, typeName)
}
