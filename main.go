package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"
)

//PBDynDNSService is a service that updates a domain using the Porkbun API
type PBDynDNSService struct {
	APIKey       string
	SecretAPIKey string
	Domain       string
	Subdomain    string
	ID           string
	FQDN         string
	IsLocal      bool
}

type PorkbunResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewPB(apikey, secretapikey, domain, subdomain string, isLocal bool) *PBDynDNSService {
	pb := &PBDynDNSService{
		APIKey:       apikey,
		SecretAPIKey: secretapikey,
		Domain:       domain,
		Subdomain:    subdomain,
		FQDN:         domain,
		IsLocal:      isLocal,
	}

	if pb.Subdomain != "" {
		pb.FQDN = fmt.Sprintf("%s.%s", subdomain, domain)
	}

	return pb
}

func main() {
	// Load the .env file if it exists
	_ = godotenv.Load()
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	// Parse command-line arguments
	daemonFlag := flag.Bool("d", false, "Run as a daemon")
	localFlag := flag.Bool("local", false, "Use local network IP")
	apiKeyFlag := flag.String("api-key", "", "Porkbun API key")
	apiSecretFlag := flag.String("api-secret", "", "Porkbun API secret")
	domainFlag := flag.String("domain", "", "Domain to update")
	subdomainFlag := flag.String("subdomain", "", "Subdomain to update")
	flag.Parse()

	// Check if values are provided as command-line arguments; otherwise, read from environment variables
	apiKey := getArgOrEnv(*apiKeyFlag, "PORKBUN_API_KEY")
	apiSecret := getArgOrEnv(*apiSecretFlag, "PORKBUN_API_SECRET")
	domain := getArgOrEnv(*domainFlag, "PBDYNDNS_DOMAIN")
	subdomain := getArgOrEnv(*subdomainFlag, "PBDYNDNS_SUBDOMAIN")
	isLocal := *localFlag || os.Getenv("PBDYNDNS_LOCAL") == "true"

	if apiKey == "" {
		log.Fatalln("No API key specified")
	}

	if apiSecret == "" {
		log.Fatalln("No API secret specified")
	}

	if domain == "" {
		domain = os.Getenv("DOMAIN")
	}

	if domain == "" {
		log.Fatalln("No domain specified")
	}

	fmt.Println("Porkbun DynDNS service started. Press Ctrl+C to exit.")

	// Check if the service should run as a daemon
	if *daemonFlag || os.Getenv("PBDYNDNS_DAEMON") == "true" {
		args := append([]string{"-d=false"}, os.Args[1:]...) // Remove the -d flag from the arguments
		cmd := exec.Command(os.Args[0], args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			log.Fatalf("Failed to start daemon: %w", err)
		}
		fmt.Printf("Running as daemon with PID %d\n", cmd.Process.Pid)
		os.Exit(0)
	}

	pb := NewPB(apiKey, apiSecret, domain, subdomain, isLocal)

	dnsRecord, err := pb.GetRecord()
	if err != nil && errors.Is(err, criticalError) {
		panic(err)
	}

	if dnsRecord != nil {
		pb.ID = dnsRecord.ID
		fmt.Printf("Found existing DNS record: %s -> %s\n", pb.FQDN, dnsRecord.Content)
	}

	// Obtain the local IP address
	ipAddr, err := pb.IP()
	if err != nil {
		panic(err)
	}

	var s string = ""
	if isLocal {
		s = "Local "
	}
	fmt.Printf("%sIP address: %s\n", s, ipAddr)

	// Update the domain using the Porkbun API
	if dnsRecord.Content == ipAddr {
		fmt.Println("Domain is already up-to-date. Skipping update.")
	} else {
		err = pb.Update(ipAddr)
		if err != nil {
			panic(err)
		}
		log.Printf("Successfully updated domain %s -> %s\n", pb.FQDN, ipAddr)
	}
	pb.Run()

	fmt.Println("Domain updated successfully")
}

// getArgOrEnv returns the value from the command-line argument if not empty, or from the environment variable otherwise
func getArgOrEnv(arg, envVar string) string {
	if arg != "" {
		return arg
	}
	return os.Getenv(envVar)
}

//IP finds the local IP address of the machine
func (pb *PBDynDNSService) IP() (string, error) {
	if !pb.IsLocal {
		resp, err := http.Get("https://api.ipify.org")
		if err != nil {
			return "", fmt.Errorf("%w: failed to get public IP address: %w", criticalError, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("%w: failed to get public IP address, HTTP status code: %d", criticalError, resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("%w: failed to read public IP address response body: %w", criticalError, err)
		}

		return string(body), nil
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}

	return "", errors.New("Local IP address not found")
}

//PorkbunRequest is the request body for the Porkbun API
type PorkbunRequest struct {
	APIKey       string `json:"apikey"`
	SecretAPIKey string `json:"secretapikey"`
	DomainRecord
}

//DomainRecord is a record to add to the domain
type DomainRecord struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      string `json:"ttl"`
	Priority string `json:"prio,omitempty"`
}

//GetRecordRequest is the request body for the Porkbun API
type GetRecordRequest struct {
	APIKey    string `json:"apikey"`
	APISecret string `json:"secretapikey"`
	Domain    string `json:"domain,omitempty"`
	Name      string `json:"name,omitempty"`
}

//criticalError is an error that should cause the program to exit
var criticalError = errors.New("serious error")

func (pb *PBDynDNSService) GetRecord() (*DomainRecord, error) {
	endpoint := fmt.Sprintf("https://porkbun.com/api/json/v3/dns/retrieve/%s", pb.Domain)

	fmt.Println("Retrieving records for domain: ", pb.Domain)
	fmt.Println(endpoint)

	apiReq := GetRecordRequest{
		APIKey:    pb.APIKey,
		APISecret: pb.SecretAPIKey,
		Domain:    pb.Domain,
		Name:      pb.Subdomain,
	}

	jsonData, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal API request JSON: %w", criticalError, err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve DNS records: %w", criticalError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: failed to retrieve DNS records, HTTP status code: %d - %w", criticalError, resp.StatusCode, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read DNS records response body: %w", criticalError, err)
	}

	var recordsResp DNSRecordsResponse
	err = json.Unmarshal(body, &recordsResp)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal DNS records response JSON: %w", criticalError, err)
	}

	if recordsResp.Status != "SUCCESS" {
		return nil, fmt.Errorf("%w: failed to retrieve DNS records, API status: %s", criticalError, recordsResp.Status)
	}

	for _, record := range recordsResp.Records {
		if record.Name == pb.FQDN && record.Type == "ALIAS" || record.Name == "CNAME" {
			err = pb.Delete(record.ID)
			if errors.Is(err, criticalError) {
				panic(err)
			}
		}
		if record.Type == "A" && record.Name == pb.FQDN {
			return &record, nil
		}
	}

	return nil, errors.New("DNS record not found")
}

func (pb *PBDynDNSService) Delete(id string) error {
	endpoint := fmt.Sprintf("https://porkbun.com/api/json/v3/dns/delete/%s/%s", pb.Domain, id)

	fmt.Println("Retrieving records for domain: ", pb.Domain)
	fmt.Println(endpoint)

	apiReq := GetRecordRequest{
		APIKey:    pb.APIKey,
		APISecret: pb.SecretAPIKey,
	}

	jsonData, err := json.Marshal(apiReq)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal API request JSON: %w", criticalError, err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("%w: failed to retrieve DNS records: %w", criticalError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: failed to retrieve DNS records, HTTP status code: %d - %w", criticalError, resp.StatusCode, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: failed to read DNS records response body: %w", criticalError, err)
	}

	var recordsResp DNSRecordsResponse
	err = json.Unmarshal(body, &recordsResp)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal DNS records response JSON: %w", criticalError, err)
	}

	if recordsResp.Status != "SUCCESS" {
		return fmt.Errorf("%w: failed to retrieve DNS records, API status: %s", criticalError, recordsResp.Status)
	}

	return nil
}

//Update updates the domain using the Porkbun API
func (pb *PBDynDNSService) Update(localIP string) error {
	var action string = "create"
	var url string = fmt.Sprintf("https://porkbun.com/api/json/v3/dns/create/%s", pb.Domain)
	if pb.ID != "" {
		action = "edit"
		url = fmt.Sprintf("https://porkbun.com/api/json/v3/dns/edit/%s/%s", pb.Domain, pb.ID)
	}
	payload := PorkbunRequest{
		APIKey:       pb.APIKey,
		SecretAPIKey: pb.SecretAPIKey,
	}

	payload.Name = pb.Subdomain
	payload.Type = "A"
	payload.Content = localIP
	payload.TTL = "600"

	if action == "edit" {
		payload.ID = pb.ID
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Println("unable to marshal request")
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Println("unable to post request")
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("unable to read body")
		return err
	}

	var porkbunResponse PorkbunResponse
	err = json.Unmarshal(body, &porkbunResponse)
	if err != nil {
		log.Println("unable to unmarshal response")
		return err
	}

	if porkbunResponse.Status != "SUCCESS" {
		return errors.New(porkbunResponse.Message)
	}

	return nil
}

type DNSRecordsResponse struct {
	Status  string         `json:"status"`
	Records []DomainRecord `json:"records"`
}

//Run runs the service in a loop, checking for IP changes every minute
func (pb *PBDynDNSService) Run() {
	dnsTicker := time.NewTicker(10 * time.Minute)

	for {
		select {
		case <-dnsTicker.C:
			ipAddr, err := pb.IP()
			if err != nil {
				log.Println(err)
				continue
			}
			record, err := pb.GetRecord()
			if err != nil {
				log.Println(err)
				continue
			}
			pb.ID = record.ID
			if err != nil || record == nil {
				log.Printf("Error getting DNS record: %v\n", err)
				err = pb.Update(ipAddr)
				continue
			}
			if record.Content != ipAddr {
				log.Printf("IP address does not match, have: %s want: %s\n", ipAddr, record.Content)
				err = pb.Update(ipAddr)
				if err != nil {
					log.Printf("Error updating domain: %v\n", err)
				} else {
					log.Printf("Domain updated successfully: %s -> %s\n", pb.FQDN, ipAddr)
				}
			}
		}
	}
}
