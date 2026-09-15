package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Color represents an RGB color with components in the [0, 1] range.
type Color struct {
	R, G, B float32
}

// UnmarshalJSON parses a hex color string such as "#rrggbb", "#rgb" or "rrggbb".
func (c *Color) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("color must be a hex string like \"#rrggbb\", got %s", string(data))
	}
	parsed, err := ParseHexColor(s)
	if err != nil {
		return err
	}
	*c = *parsed
	return nil
}

// ParseHexColor converts a hex color string to a normalized Color.
func ParseHexColor(s string) (*Color, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")

	switch len(s) {
	case 3:
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	case 6:
	case 8:
		// #rrggbbaa: the alpha channel is ignored, the shader alpha is
		// driven by the animation.
		s = s[:6]
	default:
		return nil, fmt.Errorf("invalid hex color %q: must be 3, 6 or 8 hex digits", s)
	}

	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color %q: %v", s, err)
	}

	return &Color{
		R: float32((v>>16)&0xFF) / 255.0,
		G: float32((v>>8)&0xFF) / 255.0,
		B: float32(v&0xFF) / 255.0,
	}, nil
}

// ColorConfig holds optional colors for the visual elements.
// A nil field means "keep the default animated behaviour".
type ColorConfig struct {
	Stroke           *Color `json:"stroke"`
	Particles        *Color `json:"particles"`
	Background       *Color `json:"background"`
	CursorGlow       *Color `json:"cursor_glow"`
	CursorGlowAccent *Color `json:"cursor_glow_accent"`
}

func GetColorsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".config", "hexecute")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "colors.json"), nil
}

func LoadColors() (*ColorConfig, error) {
	colorsPath, err := GetColorsPath()
	if err != nil {
		return nil, err
	}

	defaultColors := &ColorConfig{}

	data, err := os.ReadFile(colorsPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Creating default colors file at %s", colorsPath)
			if err := createDefaultColors(colorsPath, defaultColors); err != nil {
				log.Printf("Failed to create default colors file: %v", err)
			}
			return defaultColors, nil
		}
		return nil, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("Invalid colors file, using defaults: %v", err)
		return defaultColors, nil
	}

	knownKeys := getKnownKeys(ColorConfig{})
	for key := range raw {
		if !knownKeys[key] {
			log.Printf("Warning: unrecognised color key '%s' in colors file", key)
		}
	}

	colors := &ColorConfig{}
	loadColor := func(key string, dst **Color) {
		value, ok := raw[key]
		if !ok || string(value) == "null" {
			return
		}
		var c Color
		if err := json.Unmarshal(value, &c); err != nil {
			log.Printf("Warning: invalid color for '%s': %v (using default)", key, err)
			return
		}
		*dst = &c
	}
	loadColor("stroke", &colors.Stroke)
	loadColor("particles", &colors.Particles)
	loadColor("background", &colors.Background)
	loadColor("cursor_glow", &colors.CursorGlow)
	loadColor("cursor_glow_accent", &colors.CursorGlowAccent)

	return colors, nil
}

func createDefaultColors(path string, colors *ColorConfig) error {
	data, err := json.MarshalIndent(colors, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
