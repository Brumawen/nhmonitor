// nhmonitor is a tool used to monitor the status of NiceHash miner running on Windows 10.
// NiceHash Miner 2.exe tends to crash after around 8-10 hours of running.  This tool monitors the
// NiceHash webservice for the configured wallet address.  When it detects that the outstanding
// balance has not increased within the last 2 minutes, it kills the NiceHash Miner 2 process and
// restarts it.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

var myStatus Status

func main() {
	log.Println("------------------------")
	log.Println("NiceHash Monitoring Tool")
	log.Println("------------------------")

	// Load the wallet address
	add := getWalletAddress()
	if add == "" {
		// Failed to get wallet
		log.Println("Closing Tool.")
		return
	}
	log.Println("Loaded wallet address", add)

	// Check if NiceHash is running
	nh, err := isNHRunning()
	if err != nil {
		// Don't know
		log.Println("Closing Tool.")
		return
	}
	if !nh {
		// Start NiceHash
		err = startNH()
		if err != nil {
			log.Println("Closing Tool.")
			return
		}
	}

	// Start monitoring
	go monitor(add)

	// Start the web service
	http.HandleFunc("/status", webStatus)
	http.HandleFunc("/start", webStart)
	http.HandleFunc("/stop", webStop)

	http.ListenAndServe(":8080", nil)
}

func webStatus(w http.ResponseWriter, req *http.Request) {
	if time.Since(myStatus.LastCheck) > 2*time.Minute {
		myStatus.Message = "Time since last check is greater than 2 minutes!"
	} else {
		myStatus.Message = ""
	}

	b, err := json.Marshal(myStatus)
	if err != nil {
		log.Println("Failed to serialize status.")
	} else {
		w.Write(b)
	}
}

func webStart(w http.ResponseWriter, req *http.Request) {
	r, err := isNHRunning()
	if err != nil {
		w.Write([]byte("Error checking if NiceHash is running."))
	} else {
		if !r {
			startNH()
		}
		webStatus(w, req)
	}
}

func webStop(w http.ResponseWriter, req *http.Request) {
	r, err := isNHRunning()
	if err != nil {
		w.Write([]byte("Error checking if NiceHash is running."))
	} else {
		if r {
			stopNH()
		}
		webStatus(w, req)
	}
}

// monitor checks the outstanding balance, held for the configured wallet address, every 2 minutes.
// If the balance has not changed, then it assumes that NiceHash has crashed and restarts it.
func monitor(add string) {
	log.Println("Monitoring Nice Hash...")
	myStatus.Status = "Starting up"
	myStatus.LastCheck = time.Now()
	var lastBal float64
	for {
		time.Sleep(2 * time.Minute)
		myStatus.LastCheck = time.Now()
		//log.Println("Checking balance.")
		s, err := GetStats(add)
		if err != nil {
			myStatus.Status = fmt.Sprint("Error getting statistics. ", err)
		} else {
			b := s.GetBalance()
			myStatus.LastBalance = b
			if lastBal == 0 {
				// First extract
				log.Println("Balance is", b)
				myStatus.Status = "Running"
				lastBal = b
			} else if lastBal == b {
				// Balance has not changed - stopped running
				r, err := isNHRunning()
				if err != nil {
					myStatus.Status = fmt.Sprint("Error checking if NiceHash is running.", err)
				} else {
					if r {
						log.Println("Balance", b, "has not changed. NiceHash is still running. Stopping NiceHash...")
						stopNH()
					} else {
						log.Println("Balance", b, "has not changed. Starting NiceHash")
						startNH()
						lastBal = 0
					}
				}
			} else {
				// Work has been done - still running
				log.Println("NiceHash running OK. Balance is", b)
				myStatus.Status = "Running"
				lastBal = b
			}
		}
	}
}

// startNH starts NiceHash running
func startNH() error {
	myStatus.Status = "Starting NiceHash"
	log.Println("Starting NiceHash")

	u, err := user.Current()
	var fn string
	if err != nil {
		fn = "C:\\Users\\miner\\AppData\\Local\\Programs\\NiceHash Miner 2\\NiceHash Miner 2.exe"
	} else {
		fn = filepath.Join(u.HomeDir, "AppData", "Local", "Programs", "NiceHash Miner 2", "NiceHash Miner 2.exe")
	}

	if _, err := os.Stat(fn); os.IsNotExist(err) {
		log.Println("NiceHash has not been installed.")
		return err
	}
	cmd := exec.Command(fn)
	err = cmd.Start()
	if err != nil {
		log.Println("NiceHash failed to start.", err)
		return err
	}
	return nil
}

// stopNH kill the current NiceHash process
func stopNH() {
	myStatus.Status = "Stopping NiceHash"
	log.Println("Stopping NiceHash")
	cmd := exec.Command("taskkill", "/IM \"NiceHash Miner 2.exe\"")
	err := cmd.Run()
	if err != nil {
		log.Println("NiceHash failed to stop.", err)

		// Force stop
		log.Println("Forcing NiceHash to stop.")
		so, _ := exec.Command("taskkill", "/IM", "NiceHash Miner 2.exe", "/F").CombinedOutput()
		log.Println(string(so))
		log.Println("Forcing excavator to stop.")
		so, _ = exec.Command("taskkill", "/IM", "excavator.exe", "/F").CombinedOutput()
		log.Println(string(so))
	}
}

// getWalletAddress gets the wallet address from the "wallet" file.
func getWalletAddress() string {
	fn := "wallet"
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		// File does not exist
		log.Println("Wallet file does not exist.")
		return ""
	}
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Println("Error reading Wallet file.", err)
		return ""
	}
	return string(data)
}

// isNHRunning checks to see if NiceHash is currently running.
func isNHRunning() (bool, error) {
	cmd := exec.Command("tasklist")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("Error getting the process list.", err)
		return false, err
	}
	r := out.String()

	// Check if NiceHash is running
	nh := strings.Contains(r, "NiceHash Miner 2.exe")
	// Check if Excavator is running
	ex := strings.Contains(r, "excavator.exe")

	if nh && ex {
		log.Println("NiceHash is running.")
		return true, nil
	}
	log.Println("NiceHash is not running.")
	return false, nil
}
