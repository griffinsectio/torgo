/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Disconnect from tor network",
	Long:  `Disconnect from tor network and restore all configuration before connecting to tor network`,
	Run: func(cmd *cobra.Command, args []string) {
		err := os.Rename("/etc/sysctl.conf.bak", "/etc/sysctl.conf")
		if err != nil {
			log.Fatalf("Unable to restore sysctl.conf: %v", err)
		}

		exec.Command("/usr/bin/env sysctl -p /etc/sysctl.conf")

		err = os.Rename("/etc/resolv.conf.bak", "/etc/resolv.conf")
		if err != nil {
			log.Fatalf("Unable to restore resolv.conf: %v", err)
		}

		fmt.Println("stop called")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
