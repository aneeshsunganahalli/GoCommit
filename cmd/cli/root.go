/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "commit",
	Short: "CLI tool to write commit message",
	Long:  `Write a commit message with AI of your choice`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Manage LLM configuration",
}

var llmSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup your LLM provider and API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		return SetupLLM()
	},
}

var llmUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update or Delete LLM Model",
	RunE: func(cmd *cobra.Command, args []string) error {
		return UpdateLLM()
	},
}

var creatCommitMsg = &cobra.Command{
	Use: ".",
	Short: "Create Commit Message",
	RunE: func(cmd *cobra.Command, args []string) error {
		CreateCommitMsg()
		return nil
	},
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.commit-msg.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(llmCmd)
	llmCmd.AddCommand(llmSetupCmd)
	llmCmd.AddCommand(llmUpdateCmd)
	rootCmd.AddCommand(creatCommitMsg)
}

