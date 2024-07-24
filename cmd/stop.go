/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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
		if os.Geteuid() != 0 {
			fmt.Println("Please run as root")
			os.Exit(1)
		}

		fmt.Println("Restoring sysctl.conf")
		err := os.Rename("/etc/sysctl.conf.bak", "/etc/sysctl.conf")
		if err != nil {
			fmt.Printf("Unable to restore sysctl.conf: %v\n", err)
		}

		fmt.Println("Reloading sysctl.conf")
		out, err := exec.Command("sysctl", "-p", "/etc/sysctl.conf").Output()
		if err != nil {
			fmt.Printf("Could not reload sysctl.conf\n")
		}
		fmt.Println(string(out))

		fmt.Println("Restoring resolv.conf")
		err = os.Rename("/etc/resolv.conf.bak", "/etc/resolv.conf")
		if err != nil {
			fmt.Printf("Unable to restore resolv.conf: %v\n", err)
		}

		fmt.Println("Flush iptables")
		out, err = exec.Command("/usr/bin/torgo-iptables/iptables_flush.sh").Output()
		if err != nil {
			fmt.Printf("Could not flush iptables: %v\n", err)
		}
		fmt.Println(string(out))

		fmt.Println("Stopping tor")
		out, err = exec.Command("fuser", "-k", "9051/tcp").Output()
		if err != nil {
			fmt.Printf("Could not stop tor process: %v\n", err)
		}
		fmt.Print(fmt.Sprintf("process %s stopped", string(out)) + "\n")

		fmt.Println("Restart networking")
		out, err = exec.Command("/etc/init.d/networking", "restart").Output()
		if err != nil {
			fmt.Printf("Could not restart networking: %v\n", err)
		}
		fmt.Println(string(out))

		fmt.Println("Stop connecting to tor network")
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
