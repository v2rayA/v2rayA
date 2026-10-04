package service

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/v2rayA/RoutingA"
	"github.com/v2rayA/v2rayA/common/httpClient"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
)

const maxRoutingASourceBytes = 1024 * 1024

var routingAUpdate sync.Mutex
var hardcodeReplacement = regexp.MustCompile(`\$\$.+?\$\$`)

func ValidateRoutingA(text string) error {
	_, err := parseRoutingA(text)
	return err
}

func parseRoutingA(text string) (int, error) {
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = hardcodeReplacement.ReplaceAllString(lines[i], "")
	}
	parsed, err := RoutingA.Parse(strings.Join(lines, "\n"))
	if err != nil {
		return 0, fmt.Errorf("invalid RoutingA rules: %w", err)
	}
	return len(parsed), nil
}

func routingAClient(direct bool) (*http.Client, error) {
	var client *http.Client
	if direct {
		client = &http.Client{Transport: &http.Transport{
			DialContext: httpClient.DirectDialer(10*time.Second, v2ray.IsTransparentOn(configure.GetSettingNotNil())).DialContext,
		}}
	} else {
		var err error
		client, err = httpClient.GetHttpClientWithv2rayAPac()
		if err != nil {
			return nil, err
		}
	}
	client.Timeout = 30 * time.Second
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || (req.URL.Scheme != "https" && req.URL.Scheme != "http") {
			return fmt.Errorf("invalid redirect from RoutingA URL")
		}
		return nil
	}
	return client, nil
}

func ValidateRoutingASource(source configure.RoutingASource) error {
	u, err := url.Parse(strings.TrimSpace(source.URL))
	if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return fmt.Errorf("enter an HTTP or HTTPS URL to a raw RoutingA text file")
	}
	return nil
}

func FetchRoutingA(source configure.RoutingASource) (string, error) {
	if err := ValidateRoutingASource(source); err != nil {
		return "", err
	}
	u, _ := url.Parse(strings.TrimSpace(source.URL))
	client, err := routingAClient(source.DirectUpdate)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "v2rayA/"+conf.Version+" RoutingA")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not fetch RoutingA URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("RoutingA URL returned %s", resp.Status)
	}
	if resp.ContentLength > maxRoutingASourceBytes {
		return "", fmt.Errorf("RoutingA file exceeds 1 MiB")
	}
	if disposition := resp.Header.Get("Content-Disposition"); disposition != "" {
		kind, _, err := mime.ParseMediaType(disposition)
		if err != nil || kind == "attachment" {
			return "", fmt.Errorf("RoutingA URL must return raw text, not a download")
		}
	}
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || (mediaType != "text/plain" && mediaType != "application/octet-stream") {
			return "", fmt.Errorf("RoutingA URL must return a raw text file")
		}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRoutingASourceBytes+1))
	if err != nil {
		return "", err
	}
	if len(body) > maxRoutingASourceBytes {
		return "", fmt.Errorf("RoutingA file exceeds 1 MiB")
	}
	if strings.HasPrefix(http.DetectContentType(body), "text/html") {
		return "", fmt.Errorf("RoutingA URL returned an HTML page, not raw rules")
	}
	text := strings.TrimPrefix(string(body), "\ufeff")
	if !utf8.ValidString(text) || strings.ContainsRune(text, 0) || strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("RoutingA URL must contain UTF-8 rules")
	}
	rules, err := parseRoutingA(text)
	if err != nil {
		return "", err
	}
	if rules == 0 {
		return "", fmt.Errorf("invalid RoutingA rules: no rules found")
	}
	return text, nil
}

func SaveRoutingA(text string, source configure.RoutingASource) error {
	routingAUpdate.Lock()
	defer routingAUpdate.Unlock()
	var previous string
	return ApplyCoreConfig(func() func() error {
		previous = configure.GetRoutingA()
		previousSource := configure.GetRoutingASource()
		return func() error {
			if err := configure.SetRoutingA(&previous); err != nil {
				return err
			}
			return configure.SetRoutingASource(previousSource)
		}
	}, func() error {
		if err := configure.SetRoutingA(&text); err != nil {
			return err
		}
		if err := configure.SetRoutingASource(source); err != nil {
			_ = configure.SetRoutingA(&previous)
			return err
		}
		return nil
	})
}

func UpdateRoutingAFromSource() error {
	source := configure.GetRoutingASource()
	if source.URL == "" {
		return nil
	}
	text, err := FetchRoutingA(source)
	if err != nil {
		return err
	}
	routingAUpdate.Lock()
	defer routingAUpdate.Unlock()
	if configure.GetRoutingASource() != source || configure.GetRoutingA() == text {
		return nil
	}
	return ApplyCoreConfig(func() func() error {
		previous := configure.GetRoutingA()
		return func() error { return configure.SetRoutingA(&previous) }
	}, func() error { return configure.SetRoutingA(&text) })
}
