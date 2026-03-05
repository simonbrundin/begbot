package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type ProxyProvider interface {
	Name() string
	GetProxyURL() (string, error)
	GetTransport() (*http.Transport, error)
}

type NoProxy struct{}

func (p *NoProxy) Name() string                 { return "none" }
func (p *NoProxy) GetProxyURL() (string, error) { return "", nil }
func (p *NoProxy) GetTransport() (*http.Transport, error) {
	return &http.Transport{}, nil
}

type ProxyConfig struct {
	Provider string
	APIKey   string
	Country  string
	Username string
	Password string
}

func LoadFromEnv() ProxyConfig {
	return ProxyConfig{
		Provider: os.Getenv("PROXY_PROVIDER"),
		APIKey:   os.Getenv("PROXY_API_KEY"),
		Country:  os.Getenv("PROXY_COUNTRY"),
		Username: os.Getenv("PROXY_USERNAME"),
		Password: os.Getenv("PROXY_PASSWORD"),
	}
}

func NewProvider(cfg ProxyConfig) (ProxyProvider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "proxyscrape", "proxyscrape.com":
		return NewProxyScrape(cfg.APIKey, cfg.Country)
	case "proxyempire":
		return NewProxyEmpire(cfg.APIKey, cfg.Country)
	case "brightdata":
		return NewBrightData(cfg.APIKey, cfg.Country)
	case "evomi":
		return NewEvomi(cfg.APIKey, cfg.Username, cfg.Password, cfg.Country)
	case "none", "":
		return &NoProxy{}, nil
	default:
		return nil, fmt.Errorf("unknown proxy provider: %s", cfg.Provider)
	}
}

type ProxyScrape struct {
	apiKey   string
	country  string
	mu       sync.Mutex
	lastURL  string
	lastTime time.Time
}

func NewProxyScrape(apiKey, country string) (*ProxyScrape, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("ProxyScrape API key is required")
	}
	return &ProxyScrape{
		apiKey:  apiKey,
		country: country,
	}, nil
}

func (p *ProxyScrape) Name() string { return "proxyscrape" }

func (p *ProxyScrape) GetProxyURL() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if now.Sub(p.lastTime) < 10*time.Second && p.lastURL != "" {
		return p.lastURL, nil
	}

	protocol := "http"
	country := p.country
	if country == "" {
		country = "se"
	}

	p.lastURL = fmt.Sprintf("%s://%s:%s@%s.proxyscrape.com:7777",
		protocol, p.apiKey, p.country, country)

	p.lastTime = now
	return p.lastURL, nil
}

func (p *ProxyScrape) GetTransport() (*http.Transport, error) {
	proxyURL, err := p.GetProxyURL()
	if err != nil {
		return nil, err
	}

	if proxyURL == "" {
		return &http.Transport{}, nil
	}

	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		Proxy: http.ProxyURL(proxyParsed),
	}, nil
}

type ProxyEmpire struct {
	apiKey  string
	country string
	mu      sync.Mutex
}

func NewProxyEmpire(apiKey, country string) (*ProxyEmpire, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("ProxyEmpire API key is required")
	}
	return &ProxyEmpire{apiKey: apiKey, country: country}, nil
}

func (p *ProxyEmpire) Name() string { return "proxyempire" }

func (p *ProxyEmpire) GetProxyURL() (string, error) {
	country := p.country
	if country == "" {
		country = "se"
	}
	return fmt.Sprintf("http://%s:%s", country, p.apiKey), nil
}

func (p *ProxyEmpire) GetTransport() (*http.Transport, error) {
	return &http.Transport{}, nil
}

type BrightData struct {
	zone     string
	password string
	country  string
}

func NewBrightData(zone, country string) (*BrightData, error) {
	if zone == "" {
		return nil, fmt.Errorf("Bright Data zone is required")
	}
	return &BrightData{zone: zone, country: country}, nil
}

func (p *BrightData) Name() string { return "brightdata" }

func (p *BrightData) GetProxyURL() (string, error) {
	country := p.country
	if country == "" {
		country = "se"
	}
	return fmt.Sprintf("http://%s:%s", p.zone, country), nil
}

func (p *BrightData) GetTransport() (*http.Transport, error) {
	return &http.Transport{}, nil
}

type Evomi struct {
	apiKey   string
	username string
	password string
	country  string
}

func NewEvomi(apiKey, username, password, country string) (*Evomi, error) {
	if apiKey == "" && username == "" {
		return nil, fmt.Errorf("Evomi requires API key or username/password")
	}
	return &Evomi{
		apiKey:   apiKey,
		username: username,
		password: password,
		country:  country,
	}, nil
}

func (p *Evomi) Name() string { return "evomi" }

func (p *Evomi) GetProxyURL() (string, error) {
	country := p.country
	if country == "" {
		country = "se"
	}

	if p.username != "" && p.password != "" {
		return fmt.Sprintf("http://%s:%s@%s-gate.proxyserver.io:6000",
			p.username, p.password, country), nil
	}

	if p.apiKey != "" {
		return fmt.Sprintf("http://%s:%s@gate.proxyserver.io:6000",
			"session-"+p.apiKey, p.apiKey), nil
	}

	return "", fmt.Errorf("Evomi requires either username/password or API key")
}

func (p *Evomi) GetTransport() (*http.Transport, error) {
	proxyURL, err := p.GetProxyURL()
	if err != nil {
		return nil, err
	}

	if proxyURL == "" {
		return &http.Transport{}, nil
	}

	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		Proxy: http.ProxyURL(proxyParsed),
	}, nil
}
