package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func getHttpBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getDomainIP(key string, secret string, domain string, name string, recordType string) (string, error) {
	if recordType != "A" && recordType != "AAAA" {
		return "", errors.New("The recordType is neither 'A' nor 'AAAA'")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/%s/%s", domain, recordType, name), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", key, secret))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	in := make([]struct {
		Data string `json:"data"`
	}, 1)
	err = json.NewDecoder(resp.Body).Decode(&in)
	if err != nil {
		return "", err
	}
	return in[0].Data, nil
}

func putNewIP(key string, secret string, domain string, name string, recordType string, ip string) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&[1]struct {
		Data string `json:"data"`
		TTL  int64  `json:"ttl"`
	}{{
		Data: ip,
		TTL:  600,
	}})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/%s/%s", domain, recordType, name), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", key, secret))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Failed with HTTP status code %d\n", resp.StatusCode))
	}
	return nil
}

func updateIPvX(x int, key string, secret string, domain string, name string) {
	var ipProvider string
	var recordType string
	switch x {
	case 4:
		ipProvider = "https://4.ipw.cn"
		recordType = "A"
		break
	case 6:
		ipProvider = "https://6.ipw.cn"
		recordType = "AAAA"
		break
	default:
		log.Fatalln("The x is neither 4 or 6")
	}

	ip, err := getHttpBody(ipProvider)
	if err != nil {
		log.Println(err)
		return
	}
	domainIP, err := getDomainIP(key, secret, domain, name, recordType)
	if err != nil {
		log.Println(err)
		return
	}
	if domainIP == ip {
		log.Printf("The %s.%s (%s) is up to date\n", name, domain, ip)
		return
	}
	err = putNewIP(key, secret, domain, name, recordType, ip)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Successfully updated %s.%s from %s to %s\n", name, domain, domainIP, ip)
}

func main() {
	// log file flag
	logFile := flag.String("log", "", "Path for log file (will be created if it doesn't exist)")
	// required flags
	key := flag.String("key", "", "Godaddy API key")
	secret := flag.String("secret", "", "Godaddy API secret")
	domain := flag.String("domain", "", "Your top level domain (e.g., example.com) registered with Godaddy and on the same account as your API key")
	// optional flags
	name := flag.String("subdomain", "@", "The data value (aka host) for the A record. It can be a 'subdomain' (e.g., 'subdomain' where 'subdomain.example.com' is the qualified domain name). Note that such an A record must be set up first in your Godaddy account beforehand. Defaults to @. (Optional)")
	polling := flag.Int64("interval", 600, "Polling interval in seconds. Lookup Godaddy's current rate limits before setting too low. Defaults to 600. (Optional)")
	flag.Parse()

	if *logFile == "" {
		log.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Couldn't open log file: %s", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	if *secret == "" {
		log.Fatalf("You need to provide your API secret")
	}

	if *key == "" {
		log.Fatalf("You need to provide your API key")
	}

	if *domain == "" {
		log.Fatalf("You need to provide your domain")
	}

	// run
	for {
		go updateIPvX(4, *key, *secret, *domain, *name)
		go updateIPvX(6, *key, *secret, *domain, *name)
		time.Sleep(time.Second * time.Duration(*polling))
	}
}
