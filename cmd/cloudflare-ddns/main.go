package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

const (
	rtype = "A"
	ttl   = 1
	proxy = true
)

var (
	apiKey, apiEmail, oldExternalIP, externalIP, zone, name string
	oneShot                                                 bool
	updateSeconds                                           int64
)

func simpleGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyStr := string(body)

	if len(bodyStr) < 4 {
		return "", errors.New("Invalid body length")
	}

	return bodyStr, nil
}

func getExternalIP() (string, error) {
	respStr, err := simpleGet("http://ipecho.net/plain")
	if err == nil {
		return respStr, nil
	}

	respStr, err = simpleGet("http://whatismyip.akamai.com")
	if err == nil {
		return respStr, nil
	}

	respStr, err = simpleGet("http://icanhazip.com/")
	if err == nil {
		return respStr, nil
	}

	respStr, err = simpleGet("https://tnx.nl/ip")
	if err == nil {
		return respStr, nil
	}

	return "", errors.New("Could not determine external IP")
}

func updateDNS() {
	externalIP, _ = getExternalIP()

	// Construct a new API object
	api, err := cloudflare.New(apiKey, apiEmail)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName("manwe.io")
	if err != nil {
		fmt.Printf("Error updating DNS record: %#+v\n", err)
		return
	}

	// Look for an existing record
	rr := cloudflare.DNSRecord{
		Name: name + "." + zone,
	}
	records, err := api.DNSRecords(zoneID, rr)
	if err != nil {
		fmt.Printf("Error fetching DNS records: %#+v\n", err)
		return
	}

	if len(records) > 0 {
		for _, r := range records {
			if r.Type == rtype {
				rr.ID = r.ID
				rr.Type = r.Type
				rr.Content = externalIP
				rr.TTL = ttl
				rr.Proxied = proxy
				err := api.UpdateDNSRecord(zoneID, r.ID, rr)
				if err != nil {
					fmt.Printf("Error updating DNS record: %+#v\n", err)
				} else {
					fmt.Printf("Updated %s.%s -> %s\n", name, zone, externalIP)
				}
			}
		}
	} else {
		rr.Type = rtype
		rr.Content = externalIP
		rr.TTL = ttl
		rr.Proxied = proxy
		_, err = api.CreateDNSRecord(zoneID, rr)
		if err != nil {
			fmt.Printf("Error creating DNS record: %#+v\n", err)
		} else {
			fmt.Printf("Added %s.%s -> %s\n", name, zone, externalIP)
		}
	}
}

func main() {
	apiKey = os.Getenv("CF_API_KEY")
	apiEmail = os.Getenv("CF_API_EMAIL")
	zone = os.Getenv("ZONE")
	name = os.Getenv("NAME")

	oneShot = os.Getenv("ONESHOT") == "TRUE"
	updateSeconds, _ = strconv.ParseInt(os.Getenv("UPDATE_INTERVAL"), 0, 32)

	externalIP, _ = getExternalIP()
	oldExternalIP = externalIP

	updateDNS()
	if !oneShot {
		for {
			select {
			case <-time.After(time.Duration(updateSeconds) * time.Second):
				externalIP, _ = getExternalIP()
				if externalIP != oldExternalIP {
					oldExternalIP = externalIP
					updateDNS()
				} else {
					fmt.Printf("Skipping update %s == %s\n", oldExternalIP, externalIP)
				}
			}
		}
	}
}
