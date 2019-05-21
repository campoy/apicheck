// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/campoy/apicheck/apicheck/compare"
	"github.com/campoy/apicheck/apicheck/parser"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apicheck",
	Short: "Backwards compatibility check",
	Long:  `Checks whether a base and target versions of an API are backwards compatible.`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case cmd.Flag("pkg").Value.String() == "":
			return errors.New("missing package information")
		case cmd.Flag("base").Value.String() == "":
			return errors.New("missing base version")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		err := cloneAndCompare(
			cmd.Flag("pkg").Value.String(),
			cmd.Flag("base").Value.String(),
			cmd.Flag("target").Value.String(),
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func cloneAndCompare(pkg, base, target string) error {
	baseRepo, err := parser.CloneAndParse(pkg, base)
	if err != nil {
		return err
	}

	targetRepo, err := parser.CloneAndParse(pkg, target)
	if err != nil {
		return err
	}

	changes, err := compare.Repos(baseRepo, targetRepo)
	if err != nil {
		return err
	}

	for _, change := range changes {
		if !change.Compatible() {
			fmt.Println(change)
		}
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("verbose", "v", false, "Enables verbose logging mode.")
	rootCmd.Flags().StringP("pkg", "p", "", "Import path for the package to check.")
	rootCmd.Flags().StringP("base", "b", "", "The older version of the API.")
	rootCmd.Flags().StringP("target", "t", "HEAD", "The newer version of the API.")
}
