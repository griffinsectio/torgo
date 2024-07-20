/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var TOR_USER = "debian-tor"
var TOR_UID = exec.Command(fmt.Sprintf("id -ur %s", TOR_USER))

var env = `/usr/bin/env `

var sysctl = `/etc/sysctl.conf`
var torrc = `/etc/tor/torgorc`
var resolvconf = `/etc/resolv.conf`

var disable_ipv6 = `net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1`

var torrcconfig = `VirtualAddrNetwork 10.0.0.0/10
AutomapHostsOnResolve 1
TransPort 9040
DNSPort 5353
ControlPort 9051
RunAsDaemon 1`

var resolvconfig = `nameserver 127.0.0.1`

var NON_TOR = `192.168.1.0/24 192.168.0.0/24`
var iptables_rules = fmt.Sprint(`NON_TOR="%[0]s"
TOR_UID=%[1]s
TRANS_PORT="9040"

iptables -F
iptables -t nat -F

iptables -t nat -A OUTPUT -m owner --uid-owner %[1]s -j RETURN
iptables -t nat -A OUTPUT -p udp --dport 53 -j REDIRECT --to-ports 5353
for NET in $NON_TOR 127.0.0.0/9 127.128.0.0/10; do
 iptables -t nat -A OUTPUT -d $NET -j RETURN
done
iptables -t nat -A OUTPUT -p tcp --syn -j REDIRECT --to-ports "9040"

iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
for NET in $NON_TOR 127.0.0.0/8; do
iptables -A OUTPUT -d $NET -j ACCEPT
done
iptables -A OUTPUT -m owner --uid-owner %[1]s -j ACCEPT
iptables -A OUTPUT -j REJECT

iptables -A FORWARD -m string --string "BitTorrent" --algo bm --to 65535 -j DROP
iptables -A FORWARD -m string --string "BitTorrent protocol" --algo bm --to 65535 -j DROP
iptables -A FORWARD -m string --string "peer_id=" --algo bm --to 65535 -j DROP
iptables -A FORWARD -m string --string ".torrent" --algo bm --to 65535 -j DROP
iptables -A FORWARD -m string --string "announce.php?passkey=" --algo bm --to 65535 -j DROP
iptables -A FORWARD -m string --string "torrent" --algo bm --to 65535 -j DROP
iptables -A FORWARD -m string --string "announce" --algo bm --to 65535 -j DROP
iptables -A FORWARD -m string --string "info_hash" --algo bm --to 65535 -j DROP`, NON_TOR, TOR_UID)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Connect to tor network",
	Long:  `Start connecting to tor network to random country`,
	Run: func(cmd *cobra.Command, args []string) {
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
			backup, err := os.Create("/etc/sysctl.conf.bak")
			if err != nil {
				log.Fatalf("Unable to create backup file: %v", err)

			}
			backup.WriteString(string(content))

			_, err = file.WriteString(disable_ipv6)
			if err != nil {
				log.Fatalf("Unable to write to %s: %v", sysctl, err)
			}

			out, err := exec.Command("/usr/sbin/sysctl -p").Output()
			if err != nil {
				log.Fatalf("Unable to reload sysctl.conf: %v", err)
			}
			fmt.Println(string(out))

			if _, err := os.Stat(torrc); err == nil {
				content, err := os.ReadFile(torrc)
				if err != nil {
					log.Fatalf("Unable to read %s: %v", torrc, err)
				}
				if strings.Contains(string(content), torrcconfig) {
					fmt.Println("torgo configuration is already configured")
				} else {
					file, err := os.OpenFile(torrc, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
					if err != nil {
						log.Fatalf("unable to open %s: %v", torrc, err)
					}
					defer file.Close()
					file.WriteString(torrcconfig)
				}
			}

			content, err := os.ReadFile(resolvconf)
			if err != nil {
				log.Fatalf("unable to open file: %v", err)
			}
			if !strings.Contains(string(content), resolvconfig) {
				file, err := os.OpenFile(resolvconf, os.O_WRONLY|os.O_CREATE, 0644)
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

			if _, err := os.Stat("/usr/bin/tor"); err == nil {
				exec.Command(env + "systemctl stop tor")
				exec.Command("fuser -k 9051/tcp > /dev/null 2>&1")

				exec.Command(fmt.Sprintf("sudo -u %s tor -f %s > /dev/null'.format(TOR_USER, Torrc)", TOR_USER, torrc))

			} else {
				log.Fatalf("Unable to locate tor, is it installed? %v", err)
			}

			exec.Command(iptables_rules)

		} else {
			fmt.Println("IPv6 is already disabled")
		}
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
