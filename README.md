# IP Probe Sweeper

This is a simple Go program that allows to probe IPs in a given range to see which ones reply to a ping. 

To run it, provide the address range to ping in CIDR format, i.e. 1.2.0.0/16, as a flag called `cidr`. Make sure only mask bits are set in the host portion of the address.

## Flags

The possible flags are:
- cidr:  Required. The address range to ping in CIDR format, i.e. 1.2.0.0/16. Make sure only mask bits are set in the host portion of the address. 
- threads: The number of threads to use when pinging. Default is 1000.
- timeouttimeout: The amount of time to wait for a response from a host. Default is 300ms.
- verbose: Whether or not to print the results of the ping. Default is false.
- progress_freq: Frequency of printing out progress updates. Default is 1s.

## Example
```
go run prober.go --cidr="34.128.0.0/10" --threads=1200
```