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

// var TOR_UID = exec.Command(fmt.Sprintf("id -ur %s", TOR_USER))
var TOR_UID = exec.Command("id", "-ur", TOR_USER)

var env = `/usr/bin/env `

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
RunAsDaemon 1`

var resolvconfig = `nameserver 127.0.0.1`

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Connect to tor network",
	Long:  `Start connecting to tor network to random country`,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Geteuid() != 0 {
			fmt.Println("Please run as root")
			os.Exit(1)
		}

		content, err := os.ReadFile(sysctl)
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}

		file, err := os.OpenFile(sysctl, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("unable to open file: %v", err)
		}
		defer file.Close()

		if !strings.Contains(string(content), disable_ipv6) {
			fmt.Println("Disabling IPv6...")
			backup, err := os.Create(sysctl + ".bak")
			if err != nil {
				log.Fatalf("Unable to create backup file: %v", err)
			}
			backup.WriteString(string(content))

			_, err = file.WriteString(disable_ipv6)
			if err != nil {
				log.Fatalf("Unable to write to %s: %v", sysctl, err)
			}

			out, err := exec.Command("sysctl", "-p").Output()
			if err != nil {
				log.Fatalf("Unable to reload sysctl.conf: %v", err)
			}
			fmt.Println(string(out))
		} else {
			fmt.Println("IPv6 is already disabled")
		}

		_, err = os.Stat(torrc)
		if err != nil {
			fmt.Println("Creating torgorc file...")
			_, err = os.Create(torrc)
			if err != nil {
				log.Fatalf("unable to create %s: %v", torrc, err)
			}
		}
		content, err = os.ReadFile(torrc)
		if err != nil {
			log.Fatalf("Unable to read %s: %v", torrc, err)
		}
		if !strings.Contains(string(content), torrcconfig) {
			file, err := os.OpenFile(torrc, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				log.Fatalf("unable to open %s: %v", torrc, err)
			}
			defer file.Close()
			file.WriteString(torrcconfig)
		} else {
			fmt.Println("torgo configuration is already configured")
		}

		content, err = os.ReadFile(resolvconf)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("%s not exist, creating one...\n", resolvconf)

				file, err = os.OpenFile(resolvconf, os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					log.Fatalf("unable to create %s: %v", resolvconf, err)
				}
				content, err = os.ReadFile(resolvconf)
				if err != nil {
					log.Fatalf("unable to read %s: %v", resolvconf, err)
				}
				file.WriteString("")
			} else {
				log.Fatalf("unable to open file: %v", err)
			}
		}
		if !strings.Contains(string(content), resolvconfig) {
			file, err := os.OpenFile(resolvconf+".bak", os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				log.Fatalf("unable to create file %s: %v", resolvconf, err)
			}
			defer file.Close()
			file.WriteString(string(content))

			file, err = os.OpenFile(resolvconf, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				log.Fatalf("unable to open file %s: %v", resolvconf, err)
			}
			defer file.Close()

			_, err = file.WriteString(resolvconfig)
			if err != nil {
				log.Fatalf("Unable to write to %s: %v", resolvconf, err)
			}

			fmt.Println("resolv.conf is configured")
		} else {
			fmt.Println("resolv.conf already configured")
		}

		if _, err := os.Stat(tor); err == nil {
			var out, err = exec.Command("/etc/init.d/tor", "stop").Output()
			if err != nil {
				log.Fatalf("An error occurred: %v", err)
			}
			fmt.Println(string(out))

			out, _ = exec.Command("fuser", "-k", "9051/tcp").Output()
			fmt.Println(string(out))

			// out, err = exec.Command(fmt.Sprintf("sudo -u %s tor -f %s > /dev/null", TOR_USER, torrc)).Output()
			out, err = exec.Command("sudo", "-u", TOR_USER, "tor", "-f", torrc).Output()
			if err != nil {
				log.Fatalf("An error occurred: %v", err)
			}
			fmt.Println(string(out))

		} else {
			log.Fatalf("Unable to locate tor, is it installed? %v", err)
		}
		out, err := exec.Command("./iptables.sh").Output()

		if err != nil {
			log.Fatalf("An error occurred: %v", err)
		}
		fmt.Println(string(out))

		fmt.Println("Something happened!")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
