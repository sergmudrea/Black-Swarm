// Package recon implements reconnaissance modules (DNS, subdomain discovery).
package recon

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
	"github.com/miekg/dns"
)

// DNSRecon performs DNS reconnaissance on targets.
type DNSRecon struct {
	resolver string
	timeout  time.Duration
}

// NewDNSRecon creates a new DNSRecon module.
func NewDNSRecon(resolver string, timeout time.Duration) *DNSRecon {
	if resolver == "" {
		resolver = "8.8.8.8:53"
	}
	return &DNSRecon{
		resolver: resolver,
		timeout:  timeout,
	}
}

// RecordTypes lists the DNS record types to query.
var RecordTypes = []uint16{
	dns.TypeA,
	dns.TypeAAAA,
	dns.TypeMX,
	dns.TypeNS,
	dns.TypeTXT,
	dns.TypeCNAME,
	dns.TypeSOA,
}

// Scan performs DNS reconnaissance on the given domains.
func (dr *DNSRecon) Scan(ctx context.Context, domains []string) ([]protocol.Finding, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("domains must not be empty")
	}

	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
	)

	client := &dns.Client{Timeout: dr.timeout}

	for _, domain := range domains {
		for _, rtype := range RecordTypes {
			select {
			case <-ctx.Done():
				return findings, ctx.Err()
			default:
			}

			wg.Add(1)
			go func(d string, rt uint16) {
				defer wg.Done()

				msg := new(dns.Msg)
				msg.SetQuestion(dns.Fqdn(d), rt)
				msg.RecursionDesired = true

				resp, _, err := client.Exchange(msg, dr.resolver)
				if err != nil {
					return
				}

				if resp.Rcode != dns.RcodeSuccess {
					return
				}

				for _, ans := range resp.Answer {
					finding := protocol.Finding{
						Target:    d,
						Protocol:  "dns",
						Title:     fmt.Sprintf("DNS record: %s", dns.TypeToString[rt]),
						Severity:  "info",
						Timestamp: time.Now().UnixMilli(),
					}

					switch record := ans.(type) {
					case *dns.A:
						finding.Description = fmt.Sprintf("A: %s → %s", d, record.A.String())
					case *dns.AAAA:
						finding.Description = fmt.Sprintf("AAAA: %s → %s", d, record.AAAA.String())
					case *dns.MX:
						finding.Description = fmt.Sprintf("MX: %s → %s (priority %d)", d, record.Mx, record.Preference)
					case *dns.NS:
						finding.Description = fmt.Sprintf("NS: %s → %s", d, record.Ns)
					case *dns.TXT:
						finding.Description = fmt.Sprintf("TXT: %s → %s", d, strings.Join(record.Txt, " "))
					case *dns.CNAME:
						finding.Description = fmt.Sprintf("CNAME: %s → %s", d, record.Target)
					case *dns.SOA:
						finding.Description = fmt.Sprintf("SOA: %s → %s %s", d, record.Ns, record.Mbox)
					}

					mu.Lock()
					findings = append(findings, finding)
					mu.Unlock()
				}
			}(domain, rtype)
		}
	}

	wg.Wait()
	return findings, nil
}

// Resolve performs a simple DNS resolution and returns IP addresses.
func Resolve(domain string) ([]net.IP, error) {
	return net.LookupIP(domain)
}
