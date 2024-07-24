/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var TOR_USER = "debian-tor"

var TOR_UID = exec.Command("id", "-ur", TOR_USER)

var sysctl = `/etc/sysctl.conf`
var torrc = `/etc/tor/torgorc`
var resolvconf = `/etc/resolv.conf`
var tor = `/usr/bin/tor`

var disable_ipv6 = `net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1`

var torrcconfig = `VirtualAddrNetwork 10.0.0.0/10
AutomapHostsOnResolve 1
TransPort 9040
DNSPort 5353
ControlPort 9051
RunAsDaemon 1
`

var resolvconfig = `nameserver 127.0.0.1`

func CreateIfFileNotExist(filename string) {
	fmt.Printf("%s not found, creating the file...", filename)
	_, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Unable to create %s!", filename)
	}
}

func ExitIfErr(msg string, cmd *cobra.Command, args []string) {
	fmt.Println("\033[91mFatal error occurred, stop connecting to tor network...\033[37m")
	stopCmd.Run(cmd, args)
	fmt.Printf("Error details: \033[91m" + msg + "\033[37m\n")
	os.Exit(1)
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Connect to tor network",
	Long:  `Start connecting to tor network to random country`,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Geteuid() != 0 {
			fmt.Println("\033[1m\033[94mPlease run as root")
			os.Exit(1)
		}

		stopCmd.Run(cmd, args)

		// Old sysctl.conf content
		content, err := os.ReadFile(sysctl)
		if err != nil {
			if os.IsNotExist(err) {
				CreateIfFileNotExist(sysctl)
			} else {
				log.Fatalf("unable to read file: %v", err)
				stopCmd.Run(cmd, args)
			}
		}

		// Try to append to the sysctl.conf
		file, err := os.OpenFile(sysctl, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			if os.IsNotExist(err) {
				CreateIfFileNotExist(sysctl)
			} else {
				log.Fatalf("unable to read file: %v", err)
				stopCmd.Run(cmd, args)
			}
		}
		defer file.Close()

		if !strings.Contains(string(content), disable_ipv6) {
			fmt.Println("Disabling IPv6...")
			backup, err := os.Create(sysctl + ".bak")

			if err != nil {
				ExitIfErr(fmt.Sprintf("Unable to create backup file: %v", err), cmd, args)
			}

			// Write the old sysctl.conf content to the backup file
			backup.WriteString(string(content))

			_, err = file.WriteString(disable_ipv6)
			if err != nil {
				ExitIfErr(fmt.Sprintf("Unable to write to %s: %v", sysctl, err), cmd, args)
			}

			out, err := exec.Command("sysctl", "-p").Output()
			if err != nil {
				ExitIfErr(fmt.Sprintf("Unable to reload sysctl.conf: %v", err), cmd, args)
			}
			fmt.Println(string(out))
		} else {
			fmt.Println("IPv6 is already disabled")
		}

		fmt.Println("Creating torgorc file...")
		_, err = os.Create(torrc)
		if err != nil {
			ExitIfErr(fmt.Sprintf("unable to create %s: %v", torrc, err), cmd, args)
		}

		file, err = os.OpenFile(torrc, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			ExitIfErr(fmt.Sprintf("unable to open %s: %v", torrc, err), cmd, args)
		}
		defer file.Close()
		file.WriteString(torrcconfig)

		country, err := cmd.Flags().GetString("country")
		if err != nil {
			ExitIfErr(fmt.Sprintf("Unable to parse flag argument: %v", err), cmd, args)
		}

		file, err = os.OpenFile(torrc, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			ExitIfErr(fmt.Sprintf("Unable to append to %s: %v", torrc, err), cmd, args)
		}
		country = strings.ToLower(country)
		fmt.Println(country)
		file.WriteString(fmt.Sprintf("ExitNodes {%s}", country))

		content, err = os.ReadFile(resolvconf)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("%s not exist, creating one...\n", resolvconf)

				file, err = os.OpenFile(resolvconf, os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					ExitIfErr(fmt.Sprintf("unable to create %s: %v", resolvconf, err), cmd, args)
				}
				content, err = os.ReadFile(resolvconf)
				if err != nil {
					ExitIfErr(fmt.Sprintf("unable to read %s: %v", resolvconf, err), cmd, args)
				}
				file.WriteString("")
			} else {
				ExitIfErr(fmt.Sprintf("unable to open file: %v", err), cmd, args)
			}
		}
		if !strings.Contains(string(content), resolvconfig) {
			file, err := os.OpenFile(resolvconf+".bak", os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				ExitIfErr(fmt.Sprintf("unable to create file %s: %v", resolvconf, err), cmd, args)
			}
			defer file.Close()
			file.WriteString(string(content))

			file, err = os.OpenFile(resolvconf, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				ExitIfErr(fmt.Sprintf("unable to open file %s: %v", resolvconf, err), cmd, args)
			}
			defer file.Close()

			_, err = file.WriteString(resolvconfig)
			if err != nil {
				ExitIfErr(fmt.Sprintf("Unable to write to %s: %v", resolvconf, err), cmd, args)
			}

			fmt.Println("resolv.conf is configured")
		} else {
			fmt.Println("resolv.conf already configured")
		}

		if _, err := os.Stat(tor); err == nil {
			var out, err = exec.Command("/etc/init.d/tor", "stop").Output()
			if err != nil {
				ExitIfErr(fmt.Sprintf("An error occurred: %v", err), cmd, args)
			}
			fmt.Println(string(out))

			out, _ = exec.Command("fuser", "-k", "9051/tcp").Output()
			fmt.Println(string(out))

			out, err = exec.Command("sudo", "-u", TOR_USER, "tor", "-f", torrc).Output()
			if err != nil {
				ExitIfErr(fmt.Sprintf("An error occurred: %v", err), cmd, args)
			}
			fmt.Println(string(out))

		} else {
			ExitIfErr(fmt.Sprintf("Unable to locate tor, is it installed? %v", err), cmd, args)
		}
		out, err := exec.Command("/usr/bin/torgo-iptables/iptables.sh").Output()

		if err != nil {
			ExitIfErr(fmt.Sprintf("An error occurred: %v", err), cmd, args)
		}
		fmt.Println(string(out))

		fmt.Println("Connected to tor network!")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	var Country string

	startCmd.Flags().StringVarP(&Country, "country", "c", "", "Specify a country code as exit node for your connection to tor network")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
