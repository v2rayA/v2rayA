// Package conf provides a v2raya-core JSON configuration loader.
// It extends xray-core's JSON loader with support for multiObservatory,
// and automatically injects the v2ray-compatible observatory gRPC service.
// It also pre-processes custom protocols (anytls, juicity, tuic) from the JSON
// config, strips them before xray parses, then appends the built handlers.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package conf

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	obscmd "github.com/v2rayA/v2raya-core/hint/app/observatory/command"
	multiobs "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	hint_anytls "github.com/v2rayA/v2raya-core/hint/proxy/anytls"
	hint_juicity "github.com/v2rayA/v2raya-core/hint/proxy/juicity"
	hint_tuic "github.com/v2rayA/v2raya-core/hint/proxy/tuic"
	hint_tunmips "github.com/v2rayA/v2raya-core/hint/proxy/tunmips"
	xray_commander "github.com/xtls/xray-core/app/commander"
	xray_proxyman "github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/cmdarg"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/serial"
	xray_core "github.com/xtls/xray-core/core"
	xray_conf "github.com/xtls/xray-core/infra/conf"
	xray_cfgdur "github.com/xtls/xray-core/infra/conf/cfgcommon/duration"
	conf_serial "github.com/xtls/xray-core/infra/conf/serial"
	"github.com/xtls/xray-core/main/confloader"
)

// multiObsEntryJSON represents one observer group in the JSON config.
type multiObsEntryJSON struct {
	Tag             string               `json:"tag"`
	ProbeURL        string               `json:"probeURL"`
	ProbeInterval   xray_cfgdur.Duration `json:"probeInterval"`
	SubjectSelector []string             `json:"subjectSelector"`
	Settings        *obsSettingsJSON     `json:"settings"`
	Burst           *burstObsJSON        `json:"burstObservatory"`
}

// pingConfigJSON is the v5 PingConfigObject used by burstObservatory.
type pingConfigJSON struct {
	Destination   string               `json:"destination"`
	Connectivity  string               `json:"connectivity"`
	Interval      xray_cfgdur.Duration `json:"interval"`
	SamplingCount int                  `json:"samplingCount"`
	Timeout       xray_cfgdur.Duration `json:"timeout"`
	HTTPMethod    string               `json:"httpMethod"`
}

// obsSettingsJSON supports legacy observatory fields and v5 pingConfig.
type obsSettingsJSON struct {
	SubjectSelector []string             `json:"subjectSelector"`
	ProbeURL        string               `json:"probeURL"`
	ProbeInterval   xray_cfgdur.Duration `json:"probeInterval"`
	PingConfig      *pingConfigJSON      `json:"pingConfig"`
}

// burstObsJSON matches v5 burstObservatory object shape.
type burstObsJSON struct {
	SubjectSelector []string        `json:"subjectSelector"`
	PingConfig      *pingConfigJSON `json:"pingConfig"`
}

// multiObsJSON is the top-level JSON structure for the multiObservatory field.
type multiObsJSON struct {
	Observers []multiObsEntryJSON `json:"observers"`
}

// extendedJSON captures v2raya-core extension fields alongside standard xray JSON.
type extendedJSON struct {
	MultiObservatory *multiObsJSON `json:"multiObservatory"`
	BurstObservatory *burstObsJSON `json:"burstObservatory"`
}

// customOutboundJSON is the raw JSON form of an anytls or juicity outbound.
type customOutboundJSON struct {
	Protocol       string          `json:"protocol"`
	Tag            string          `json:"tag"`
	Settings       json.RawMessage `json:"settings"`
	StreamSettings json.RawMessage `json:"streamSettings"`
}

// customInboundJSON is the raw JSON form of a tun-mips inbound.
type customInboundJSON struct {
	Protocol string                    `json:"protocol"`
	Tag      string                    `json:"tag"`
	Settings json.RawMessage           `json:"settings"`
	Sniffing *xray_conf.SniffingConfig `json:"sniffing"`
}

// customProtocols is the set of outbound protocols handled by hint/proxy.
var customProtocols = map[string]bool{
	"anytls":  true,
	"juicity": true,
	"tuic":    true,
}

// customInboundProtocols is the set of inbound protocols handled by
// hint/proxy. It is kept apart from customProtocols so that a protocol used
// in the wrong direction is rejected instead of being stripped and then
// silently dropped.
var customInboundProtocols = map[string]bool{
	"tun-mips": true,
}

// customConfig is everything stripped from one JSON document before xray
// parses it.
type customConfig struct {
	outbounds []customOutboundJSON
	inbounds  []customInboundJSON
}

func (c *customConfig) empty() bool {
	return c == nil || (len(c.outbounds) == 0 && len(c.inbounds) == 0)
}

func (c *customConfig) merge(other *customConfig) {
	if other == nil {
		return
	}
	c.outbounds = append(c.outbounds, other.outbounds...)
	c.inbounds = append(c.inbounds, other.inbounds...)
}

// stripCustom removes hint/proxy outbounds and inbounds from raw JSON and
// returns the modified document with the stripped descriptors. Invalid JSON
// or a document without those arrays is returned unchanged, nothing
// stripped, for xray to report on.
func stripCustom(raw []byte) ([]byte, *customConfig, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		return raw, nil, nil
	}
	custom := &customConfig{}
	changed := false

	if outboundsRaw, ok := doc["outbounds"]; ok {
		var outbounds []json.RawMessage
		if err := json.Unmarshal(outboundsRaw, &outbounds); err == nil {
			var remaining []json.RawMessage
			for _, ob := range outbounds {
				var probe struct {
					Protocol string `json:"protocol"`
					Tag      string `json:"tag"`
				}
				if err := json.Unmarshal(ob, &probe); err != nil {
					remaining = append(remaining, ob)
					continue
				}
				if customInboundProtocols[probe.Protocol] {
					return raw, nil, errors.New("protocol ", probe.Protocol, " is an inbound, not valid as outbound (tag ", probe.Tag, ")")
				}
				if !customProtocols[probe.Protocol] {
					remaining = append(remaining, ob)
					continue
				}
				var c customOutboundJSON
				if err := json.Unmarshal(ob, &c); err != nil {
					return raw, nil, errors.New("invalid ", probe.Protocol, " outbound (tag ", probe.Tag, ")").Base(err)
				}
				custom.outbounds = append(custom.outbounds, c)
			}
			if len(custom.outbounds) > 0 {
				newOutbounds, err := json.Marshal(remaining)
				if err != nil {
					return raw, nil, err
				}
				doc["outbounds"] = newOutbounds
				changed = true
			}
		}
	}

	if inboundsRaw, ok := doc["inbounds"]; ok {
		var inbounds []json.RawMessage
		if err := json.Unmarshal(inboundsRaw, &inbounds); err == nil {
			var remaining []json.RawMessage
			for _, ib := range inbounds {
				var probe struct {
					Protocol string `json:"protocol"`
					Tag      string `json:"tag"`
				}
				if err := json.Unmarshal(ib, &probe); err != nil {
					remaining = append(remaining, ib)
					continue
				}
				if customProtocols[probe.Protocol] {
					return raw, nil, errors.New("protocol ", probe.Protocol, " is an outbound, not valid as inbound (tag ", probe.Tag, ")")
				}
				if !customInboundProtocols[probe.Protocol] {
					remaining = append(remaining, ib)
					continue
				}
				var c customInboundJSON
				if err := json.Unmarshal(ib, &c); err != nil {
					return raw, nil, errors.New("invalid ", probe.Protocol, " inbound (tag ", probe.Tag, ")").Base(err)
				}
				custom.inbounds = append(custom.inbounds, c)
			}
			if len(custom.inbounds) > 0 {
				newInbounds, err := json.Marshal(remaining)
				if err != nil {
					return raw, nil, err
				}
				doc["inbounds"] = newInbounds
				changed = true
			}
		}
	}

	if !changed {
		return raw, nil, nil
	}
	modified, err := json.Marshal(doc)
	if err != nil {
		return raw, nil, err
	}
	return modified, custom, nil
}

// buildCustomInbounds converts stripped inbound descriptors into
// InboundHandlerConfig entries. The receiver settings carry the sniffing
// block so the proxyman builds the same sniffing context it would for a
// native inbound; there is no port because the device is the input.
func buildCustomInbounds(customs []customInboundJSON) ([]*xray_core.InboundHandlerConfig, error) {
	result := make([]*xray_core.InboundHandlerConfig, 0, len(customs))
	for _, c := range customs {
		receiver := &xray_proxyman.ReceiverConfig{}
		if c.Sniffing != nil {
			sniffing, err := c.Sniffing.Build()
			if err != nil {
				return nil, errors.New("invalid sniffing settings for tag ", c.Tag).Base(err)
			}
			receiver.SniffingSettings = sniffing
		}
		var proxySettings *serial.TypedMessage
		switch c.Protocol {
		case "tun-mips":
			var raw hint_tunmips.ConfigJSON
			if len(c.Settings) == 0 {
				return nil, errors.New("tun-mips inbound (tag ", c.Tag, ") has no settings")
			}
			if err := json.Unmarshal(c.Settings, &raw); err != nil {
				return nil, errors.New("invalid tun-mips settings for tag ", c.Tag).Base(err)
			}
			cfg, err := raw.Build()
			if err != nil {
				return nil, errors.New("invalid tun-mips settings for tag ", c.Tag).Base(err)
			}
			proxySettings = serial.ToTypedMessage(cfg)
		default:
			return nil, errors.New("unknown custom inbound protocol ", c.Protocol)
		}
		result = append(result, &xray_core.InboundHandlerConfig{
			Tag:              c.Tag,
			ReceiverSettings: serial.ToTypedMessage(receiver),
			ProxySettings:    proxySettings,
		})
	}
	return result, nil
}

// appendCustom adds the stripped descriptors to a built core config.
func appendCustom(coreConfig *xray_core.Config, custom *customConfig) error {
	if custom.empty() {
		return nil
	}
	if len(custom.outbounds) > 0 {
		customOutbounds, err := buildCustomOutbounds(custom.outbounds)
		if err != nil {
			return err
		}
		coreConfig.Outbound = append(coreConfig.Outbound, customOutbounds...)
	}
	if len(custom.inbounds) > 0 {
		customInbounds, err := buildCustomInbounds(custom.inbounds)
		if err != nil {
			return err
		}
		coreConfig.Inbound = append(coreConfig.Inbound, customInbounds...)
	}
	return nil
}

// buildCustomOutbounds converts stripped custom outbound JSON descriptors into
// xray OutboundHandlerConfig proto messages ready to append to coreConfig.Outbound.
func buildCustomOutbounds(customs []customOutboundJSON) ([]*xray_core.OutboundHandlerConfig, error) {
	result := make([]*xray_core.OutboundHandlerConfig, 0, len(customs))
	for _, c := range customs {
		senderSettings, err := buildCustomSenderSettings(c.StreamSettings)
		if err != nil {
			return nil, errors.New("invalid stream settings for tag ", c.Tag).Base(err)
		}

		var proxySettings *serial.TypedMessage
		switch c.Protocol {
		case "anytls":
			var cfg hint_anytls.ClientConfig
			if c.Settings != nil {
				if err := json.Unmarshal(c.Settings, &cfg); err != nil {
					return nil, errors.New("invalid anytls settings for tag ", c.Tag).Base(err)
				}
			}
			proxySettings = serial.ToTypedMessage(&cfg)
		case "juicity":
			var cfg hint_juicity.ClientConfig
			if c.Settings != nil {
				if err := json.Unmarshal(c.Settings, &cfg); err != nil {
					return nil, errors.New("invalid juicity settings for tag ", c.Tag).Base(err)
				}
			}
			proxySettings = serial.ToTypedMessage(&cfg)
		case "tuic":
			var cfg hint_tuic.ClientConfig
			if c.Settings != nil {
				if err := json.Unmarshal(c.Settings, &cfg); err != nil {
					return nil, errors.New("invalid tuic settings for tag ", c.Tag).Base(err)
				}
			}
			proxySettings = serial.ToTypedMessage(&cfg)
		default:
			continue
		}
		result = append(result, &xray_core.OutboundHandlerConfig{
			Tag:            c.Tag,
			SenderSettings: senderSettings,
			ProxySettings:  proxySettings,
		})
	}
	return result, nil
}

func buildCustomSenderSettings(streamSettings json.RawMessage) (*serial.TypedMessage, error) {
	if len(streamSettings) == 0 {
		return serial.ToTypedMessage(&xray_proxyman.SenderConfig{}), nil
	}

	type outboundJSON struct {
		Protocol       string          `json:"protocol"`
		Tag            string          `json:"tag"`
		StreamSettings json.RawMessage `json:"streamSettings"`
	}
	raw, err := json.Marshal(struct {
		Outbounds []outboundJSON `json:"outbounds"`
	}{
		Outbounds: []outboundJSON{{
			Protocol:       "freedom",
			Tag:            "v2raya-native-sender",
			StreamSettings: streamSettings,
		}},
	})
	if err != nil {
		return nil, err
	}

	config, err := conf_serial.DecodeJSONConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	coreConfig, err := config.Build()
	if err != nil {
		return nil, err
	}
	if len(coreConfig.Outbound) != 1 || coreConfig.Outbound[0].SenderSettings == nil {
		return nil, errors.New("failed to build sender settings")
	}
	return coreConfig.Outbound[0].SenderSettings, nil
}

// injectMultiObservatory appends the multiobservatory.Config TypedMessage to coreConfig.App.
func injectMultiObservatory(coreConfig *xray_core.Config, mo *multiObsJSON) {
	cfg := &multiobs.Config{}
	for _, e := range mo.Observers {
		observer := normalizeObserverConfig(e)
		if observer == nil {
			continue
		}
		cfg.Observers = append(cfg.Observers, observer)
	}
	if len(cfg.Observers) == 0 {
		return
	}
	coreConfig.App = append(coreConfig.App, serial.ToTypedMessage(cfg))
}

// injectBurstObservatoryAsMulti adapts v5 burstObservatory to a single-group MultiObservatory.
// This keeps v2rayA's per-tag API path working: unknown tag requests fall back to aggregated result.
func injectBurstObservatoryAsMulti(coreConfig *xray_core.Config, burst *burstObsJSON) {
	if burst == nil || len(burst.SubjectSelector) == 0 {
		return
	}
	observer := &multiobs.ObserverConfig{
		Tag:             "_burst_global",
		SubjectSelector: burst.SubjectSelector,
	}
	if burst.PingConfig != nil {
		observer.ProbeUrl = burst.PingConfig.Destination
		observer.ProbeInterval = int64(burst.PingConfig.Interval)
	}
	cfg := &multiobs.Config{Observers: []*multiobs.ObserverConfig{observer}}
	coreConfig.App = append(coreConfig.App, serial.ToTypedMessage(cfg))
}

// normalizeObserverConfig converts legacy + v5 observatory JSON into one runtime observer config.
func normalizeObserverConfig(e multiObsEntryJSON) *multiobs.ObserverConfig {
	subjectSelector := e.SubjectSelector
	probeURL := e.ProbeURL
	probeInterval := e.ProbeInterval

	if e.Settings != nil {
		if len(subjectSelector) == 0 {
			subjectSelector = e.Settings.SubjectSelector
		}
		if probeURL == "" {
			probeURL = e.Settings.ProbeURL
		}
		if probeInterval == 0 {
			probeInterval = e.Settings.ProbeInterval
		}
		if e.Settings.PingConfig != nil {
			if probeURL == "" {
				probeURL = e.Settings.PingConfig.Destination
			}
			if probeInterval == 0 {
				probeInterval = e.Settings.PingConfig.Interval
			}
		}
	}

	if e.Burst != nil {
		if len(subjectSelector) == 0 {
			subjectSelector = e.Burst.SubjectSelector
		}
		if e.Burst.PingConfig != nil {
			if probeURL == "" {
				probeURL = e.Burst.PingConfig.Destination
			}
			if probeInterval == 0 {
				probeInterval = e.Burst.PingConfig.Interval
			}
		}
	}

	if len(subjectSelector) == 0 {
		return nil
	}

	return &multiobs.ObserverConfig{
		Tag:             e.Tag,
		ProbeUrl:        probeURL,
		ProbeInterval:   int64(probeInterval),
		SubjectSelector: subjectSelector,
	}
}

// reorderAppsForAPIReadiness ensures the commander (API/gRPC) app starts before
// the proxyman inbound manager. By default, xray puts InboundConfig first in
// config.App, which means the API port becomes connectable before the gRPC server
// is running — causing "did not connect: context deadline exceeded" from v2rayA's
// ObservatoryProducer. Moving the commander before InboundConfig fixes this race.
func reorderAppsForAPIReadiness(coreConfig *xray_core.Config) {
	const commanderType = "xray.app.commander.Config"
	const inboundMgrType = "xray.app.proxyman.InboundConfig"

	commanderIdx := -1
	inboundIdx := -1
	for i, app := range coreConfig.App {
		switch app.Type {
		case commanderType:
			commanderIdx = i
		case inboundMgrType:
			inboundIdx = i
		}
	}
	// Only reorder if commander exists and is after the inbound manager.
	if commanderIdx <= inboundIdx || commanderIdx < 0 || inboundIdx < 0 {
		return
	}
	commander := coreConfig.App[commanderIdx]
	newApp := make([]*serial.TypedMessage, 0, len(coreConfig.App))
	for i, app := range coreConfig.App {
		if i == inboundIdx {
			newApp = append(newApp, commander)
		}
		if i != commanderIdx {
			newApp = append(newApp, app)
		}
	}
	coreConfig.App = newApp
}

// injectObsCommandService adds the v2ray-compatible ObservatoryService to the commander's

// service list. If no commander is found (no api section), this is a no-op.
func injectCompatService(coreConfig *xray_core.Config) {
	const commanderType = "xray.app.commander.Config"
	for i, app := range coreConfig.App {
		if app.Type != commanderType {
			continue
		}
		msg, err := app.GetInstance()
		if err != nil {
			continue
		}
		cmdCfg, ok := msg.(*xray_commander.Config)
		if !ok {
			continue
		}
		cmdCfg.Service = append(cmdCfg.Service, serial.ToTypedMessage(&obscmd.Config{}))
		coreConfig.App[i] = serial.ToTypedMessage(cmdCfg)
		return
	}
}

// loadAndExtend reads a config file, decodes it as xray conf.Config, and
// also extracts the extendedJSON fields (e.g. multiObservatory).
// Custom-protocol outbounds (anytls, juicity) are stripped from the JSON
// before xray processes it, and returned separately.
func loadAndExtend(arg string) (*xray_conf.Config, *extendedJSON, *customConfig, error) {
	r, err := confloader.LoadConfig(arg)
	if err != nil {
		return nil, nil, nil, errors.New("failed to read config: ", arg).Base(err)
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, nil, errors.New("failed to read config bytes: ", arg).Base(err)
	}
	modified, customs, err := stripCustom(raw)
	if err != nil {
		return nil, nil, nil, errors.New("failed to strip custom protocols: ", arg).Base(err)
	}
	c, err := conf_serial.DecodeJSONConfig(bytes.NewReader(modified))
	if err != nil {
		return nil, nil, nil, errors.New("failed to decode config: ", arg).Base(err)
	}
	ext := &extendedJSON{}
	// Ignore JSON errors here — unknown fields in extendedJSON are fine.
	_ = json.Unmarshal(raw, ext)
	return c, ext, customs, nil
}

// buildConfigFromFiles is our replacement for xray's serial.BuildConfig.
// It strips anytls/juicity outbounds before xray's infra/conf processes them,
// then appends them as manually-built OutboundHandlerConfig entries.
func buildConfigFromFiles(files []*xray_core.ConfigSource) (*xray_core.Config, error) {
	cf := &xray_conf.Config{}
	var ext *extendedJSON
	allCustoms := &customConfig{}

	for i, file := range files {
		errors.LogInfo(context.Background(), "v2raya-core: reading config: ", file)
		r, err := confloader.LoadConfig(file.Name)
		if err != nil {
			return nil, errors.New("failed to read config: ", file).Base(err)
		}
		raw, err := io.ReadAll(r)
		if err != nil {
			return nil, errors.New("failed to read config bytes: ", file).Base(err)
		}

		// Strip custom protocols before passing to xray.
		modified, customs, err := stripCustom(raw)
		if err != nil {
			return nil, errors.New("failed to strip custom protocols: ", file).Base(err)
		}
		allCustoms.merge(customs)

		c, err := conf_serial.DecodeJSONConfig(bytes.NewReader(modified))
		if err != nil {
			return nil, errors.New("failed to decode config: ", file).Base(err)
		}
		if i == 0 {
			*cf = *c
		} else {
			cf.Override(c, file.Name)
		}

		e := &extendedJSON{}
		_ = json.Unmarshal(raw, e)
		if e.MultiObservatory != nil || e.BurstObservatory != nil {
			ext = e
		}
	}

	coreConfig, err := cf.Build()
	if err != nil {
		return nil, err
	}
	if ext != nil && ext.MultiObservatory != nil {
		injectMultiObservatory(coreConfig, ext.MultiObservatory)
	} else if ext != nil && ext.BurstObservatory != nil {
		injectBurstObservatoryAsMulti(coreConfig, ext.BurstObservatory)
	}
	injectCompatService(coreConfig)
	reorderAppsForAPIReadiness(coreConfig)
	if err := appendCustom(coreConfig, allCustoms); err != nil {
		return nil, err
	}
	return coreConfig, nil
}

func init() {
	// Override ConfigBuilderForFiles so that when core.LoadConfig is called with cmdarg.Arg,
	// custom protocols (anytls, juicity) are stripped before xray's infra/conf processes them,
	// then appended as fully built OutboundHandlerConfig entries.
	xray_core.ConfigBuilderForFiles = buildConfigFromFiles

	common.Must(xray_core.RegisterConfigLoader(&xray_core.ConfigFormat{
		Name:      "JSON",
		Extension: []string{"json"},
		Loader: func(input interface{}) (*xray_core.Config, error) {
			var ext *extendedJSON
			allCustoms := &customConfig{}

			switch v := input.(type) {
			case cmdarg.Arg:
				cf := &xray_conf.Config{}
				for i, arg := range v {
					errors.LogInfo(context.Background(), "v2raya-core: reading config: ", arg)
					c, e, customs, err := loadAndExtend(arg)
					if err != nil {
						return nil, err
					}
					if i == 0 {
						*cf = *c
					} else {
						cf.Override(c, arg)
					}
					// Use extensions from the last file that defines them.
					if e != nil && (e.MultiObservatory != nil || e.BurstObservatory != nil) {
						ext = e
					}
					allCustoms.merge(customs)
				}
				coreConfig, err := cf.Build()
				if err != nil {
					return nil, err
				}
				if ext != nil && ext.MultiObservatory != nil {
					injectMultiObservatory(coreConfig, ext.MultiObservatory)
				} else if ext != nil && ext.BurstObservatory != nil {
					injectBurstObservatoryAsMulti(coreConfig, ext.BurstObservatory)
				}
				injectCompatService(coreConfig)
				reorderAppsForAPIReadiness(coreConfig)
				if err := appendCustom(coreConfig, allCustoms); err != nil {
					return nil, err
				}
				return coreConfig, nil

			case io.Reader:
				raw, err := io.ReadAll(v)
				if err != nil {
					return nil, errors.New("failed to read config reader").Base(err)
				}
				modified, customs, err := stripCustom(raw)
				if err != nil {
					return nil, errors.New("failed to strip custom protocols").Base(err)
				}
				c, err := conf_serial.DecodeJSONConfig(bytes.NewReader(modified))
				if err != nil {
					return nil, errors.New("failed to decode JSON config").Base(err)
				}
				e := &extendedJSON{}
				_ = json.Unmarshal(raw, e)
				coreConfig, err := c.Build()
				if err != nil {
					return nil, err
				}
				if e.MultiObservatory != nil {
					injectMultiObservatory(coreConfig, e.MultiObservatory)
				} else if e.BurstObservatory != nil {
					injectBurstObservatoryAsMulti(coreConfig, e.BurstObservatory)
				}
				injectCompatService(coreConfig)
				reorderAppsForAPIReadiness(coreConfig)
				if err := appendCustom(coreConfig, customs); err != nil {
					return nil, err
				}
				return coreConfig, nil

			default:
				return nil, errors.New("unknown config input type")
			}
		},
	}))
}
