package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slices"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

type SubdomainsExtractor struct{}

type SubdomainsDataHolder struct {
	CommonName string `json:"common_name"`
	NameValue  string `json:"name_value"`
}

func (se SubdomainsExtractor) isValidDomain(domain string) bool {
	pattern := `^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`
	return regexp.MustCompile(pattern).MatchString(domain)
}

func (se SubdomainsExtractor) getFromCertificate(domain string) []string {
	var holders []SubdomainsDataHolder
	var subdomains []string
	if !se.isValidDomain(domain) {
		log.Fatal("Incompatible domain name. Please retry.")
	}
	requestURL := fmt.Sprintf("https://crt.sh/?q=%s&output=json", domain)
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		if Body.Close() != nil {
			log.Fatal(err)
		}
	}(resp.Body)
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}
	if err = json.Unmarshal(body, &holders); err != nil {
		log.Panic(err)
	}
	// Transform to []string
	for _, h := range holders {
		nameValues := append(strings.Split(h.NameValue, "\n"), h.CommonName)
		for _, val := range nameValues {
			if !slices.Contains(subdomains, val) {
				subdomains = append(subdomains, val)
			}
		}
	}
	sort.Strings(subdomains)
	return subdomains
}

func main() {
	se := SubdomainsExtractor{}
	if len(os.Args) <= 1 {
		log.Panic("Wrong number of arguments!")
	}
	fmt.Println()
	targetDomain := os.Args[1]
	for _, subdomain := range se.getFromCertificate(targetDomain) {
		log.Println("[!]", subdomain)
	}
}
