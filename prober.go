// Prober is a package that allows for multithreading attempting pings.
package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/seancfoley/ipaddress-go/ipaddr"
	"k8s.io/klog/v2"
)

// Const for how many threads to use.
const numThreads = 1000

// Prober is a struct that contains the information needed to probe a host.
type Prober struct {
	addrCh chan string
	// Timeout is the amount of time to wait for a response from a host.
	Timeout int
	// Threads is the number of threads to use when probing.
	Threads int
	// Verbose is a flag that determines whether or not to print the results of the probe.
	Verbose bool

	CountReachable   int
	CountUnreachable int
}

// NewProber creates a new Prober struct with the given hosts, timeout, and threads.
func NewProber(addrCh chan string, timeout int, threads int, verbose bool) *Prober {
	return &Prober{addrCh, timeout, threads, verbose, 0, 0}
}

// Probe attempts to ping all of the hosts in the Prober struct.
func (p *Prober) Probe() {
	var wg sync.WaitGroup
	// Create a buffered channel to hold the number of threads.
	// This will allow us to limit the number of threads running at a time.
	threads := make(chan struct{}, p.Threads)

	// Select out addresses from the addr Channel one by one.
	// For each address, add a thread to the channel and start a goroutine to ping the host.
	for host := range p.addrCh {
		// Add a thread to the channel.
		threads <- struct{}{}
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			// Attempt to ping the host.
			// If the host is not pingable, the output will be an empty string.
			output := ping(host, p.Timeout)
			// If the output is not empty, the host is pingable.
			if output != "" {
				if p.Verbose {
					fmt.Println(host + " is pingable")
				}
				p.CountReachable++
			} else {
				if p.Verbose {
					fmt.Println(host + " is not pingable")
				}
				p.CountUnreachable++
			}
			// Remove a thread from the channel.
			<-threads
		}(host)
	}
	wg.Wait()
}

// ping attempts to ping the given host.
// If the host is pingable, the output will be a string containing the output of the ping command.
// If the host is not pingable, the output will be an empty string.
func ping(host string, timeout int) string {
	// Create the command to ping the host.
	cmd := exec.Command("ping", "-c", "1", "-W", strconv.Itoa(timeout), host)
	// Run the command.
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// Convert the output to a string.
	outputString := string(output)
	// If the output contains the host, the host is pingable.
	if strings.Contains(outputString, host) {
		return outputString
	}
	// Otherwise, the host is not pingable.
	return ""
}

// GetIPsFromCIDR takes a CIDR string and a channel and sends all of the IP addresses in the CIDR to the channel.
func GetIPsFromCIDR(cidr string, addrCh chan string) {
	defer close(addrCh)

	block := ipaddr.NewIPAddressString(cidr).GetAddress()
	for i := block.Iterator(); i.HasNext(); {
		addr := i.Next()
		addrCh <- strings.Split(addr.ToNormalizedString(), "/")[0]
	}
}

func (p *Prober) ReportProgress() {
	fmt.Printf("%.2f (Reachable: %d, Unreachable: %d)\n", float64(p.CountReachable)/float64(p.CountReachable+p.CountUnreachable)*100, p.CountReachable, p.CountUnreachable)
}

// main is the entry point of the program.
func main() {
	// Turn off klog info logs.
	klog.SetOutputBySeverity("INFO", ioutil.Discard)

	// Make a channel to hold the IP addresses. Buffer up to 1000 addresses.
	addrCh := make(chan string, numThreads)

	// IPs of interest: 34.128.0.0/10
	// Genereate all of the IP addresses in the CIDR and send them to the channel.
	go GetIPsFromCIDR("34.133.0.0/16", addrCh)

	// Create a new Prober struct.
	prober := NewProber(addrCh, 1, numThreads, false)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			prober.ReportProgress()
		}
	}()
	// Probe the hosts.
	prober.Probe()
	prober.ReportProgress()
}
