package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

var (
	zoneData  = make(map[string][]dns.RR)
	staleData = make(map[string][]dns.RR)
)

// CustomHandler handles DNS requests
type CustomHandler struct{}

func (h *CustomHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	// Check if the request is for an A record
	if len(r.Question) > 0 && r.Question[0].Qtype == dns.TypeA {
		question := r.Question[0].Name
		labels := dns.SplitDomainName(question)
		domain := strings.Join(labels[len(labels)-2:], ".")

		// Check if the zone has any records
		if records, ok := zoneData[domain]; ok {
			m.Answer = append(m.Answer, records...)
		} else if records, ok := staleData[domain]; ok {
			m.Answer = append(m.Answer, records...)
		}
	}

	w.WriteMsg(m)
}

func main() {
	var (
		listenAddr        = "127.0.0.1:53"
		consoleCmd        = "flush"
		flushRecordCmd    = "flushrecord"
		flushStaleDataCmd = "flushstale"
		apiEndpoint       = "http://localhost:8080/dns"
		apiUsername       = "your-username"
		apiPassword       = "your-password"
		apiRequestTimeout = 5 * time.Second
	)

	// Load zone data from the API
	err := loadZoneDataFromAPI(apiEndpoint, apiUsername, apiPassword, apiRequestTimeout)
	if err != nil {
		log.Fatalf("Failed to load zone data from API: %s\n", err.Error())
	}

	// Start the DNS server
	server := &dns.Server{Addr: listenAddr, Net: "udp"}
	server.Handler = &CustomHandler{}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start DNS server: %s\n", err.Error())
		}
	}()

	log.Printf("DNS server started and listening on %s\n", listenAddr)

	// Handle cache flush and record flush commands from console
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if text == consoleCmd {
				log.Println("Cache flush command received. Flushing stale data.")
				staleData = make(map[string][]dns.RR)
			} else if text == flushRecordCmd {
				log.Println("Flush record command received. Please provide the record to flush (e.g., example.com A).")
				record, _ := reader.ReadString('\n')
				record = strings.TrimSpace(record)
				flushRecord(record)
			} else if text == flushStaleDataCmd {
				log.Println("Flush stale data command received. Flushing all stale records.")
				flushStaleData()
			}
			// Add additional commands or logic as needed
		}
	}()

	// Keep the main goroutine running
	<-sig
	log.Println("Shutting down the DNS server...")
	server.Shutdown()
}

// flushRecord removes the specified record from zoneData and moves it to staleData
func flushRecord(record string) {
	fields := strings.Fields(record)
	if len(fields) < 2 {
		log.Println("Invalid record format. Please provide the record in the format 'domain type' (e.g., example.com A).")
		return
	}

	domain := fields[0]
	rrType := dns.StringToType[strings.ToUpper(fields[1])]

	if records, ok := zoneData[domain]; ok {
		var updatedRecords []dns.RR
		for _, r := range records {
			if r.Header().Rrtype != rrType {
				updatedRecords = append(updatedRecords, r)
			} else {
				staleData[domain] = append(staleData[domain], r)
				log.Printf("Record %s flushed and moved to stale data\n", r.String())
			}
		}
		zoneData[domain] = updatedRecords
	} else {
		log.Printf("No records found for domain %s\n", domain)
	}
}

// flushStaleData clears all stale records from staleData
func flushStaleData() {
	staleData = make(map[string][]dns.RR)
	log.Println("Stale data flushed.")
}

// loadZoneDataFromAPI fetches the zone data from the API and populates the zoneData map
func loadZoneDataFromAPI(endpoint, username, password string, timeout time.Duration) error {
	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to fetch zone data from API")
	}

	var zoneRecords []dns.RR
	err = json.NewDecoder(resp.Body).Decode(&zoneRecords)
	if err != nil {
		return err
	}

	// Group records by domain
	zoneData = make(map[string][]dns.RR)
	for _, record := range zoneRecords {
		if record.Header().Rrtype == dns.TypeA {
			labels := dns.SplitDomainName(record.Header().Name)
			domain := strings.Join(labels[len(labels)-2:], ".")
			zoneData[domain] = append(zoneData[domain], record)
		}
	}

	return nil
}
