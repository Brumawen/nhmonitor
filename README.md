# nhmonitor
NiceHash monitoring tool written in go.

nhmonitor is a tool used to monitor the status of NiceHash miner running on Windows 10.
NiceHash Miner 2.exe tends to crash after around 8-10 hours of running.  This tool monitors the
NiceHash webservice for the configured wallet address.  When it detects that the outstanding
balance has not increased within the last 2 minutes, it kills the NiceHash Miner 2 process and
restarts it.

This monitoring tool is for Windows 10 only for now.

## Installation

`go get github.com/brumawen/nhmonitor`

## Setup
* build nhmonitor
* create a file with the name "wallet" in the same folder as nhmonitor.exe
* save the wallet address to this file.
* run nhmonitor