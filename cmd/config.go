package main

import (
	"errors"
	"strconv"
	"strings"
)

type SortMode string

const (
	SortByTime   SortMode = "time"
	SortByName   SortMode = "name"
	SortByAmount SortMode = "amount"
)

type Config struct {
	UserID                       string
	APIToken                     string
	Output                       string
	TotalSponsor                 int
	AvatarSize                   int
	Margin                       int
	AvatarsPerRow                int
	FontFile                     string
	FontSizeScale                int
	PaddingXScale                int
	PaddingYScale                int
	Sort                         SortMode
	AnimationDelay               float32
	ActiveSponsorOpacity         float32
	ExpiredSponsorOpacity        float32
	UseActiveOpacityWhenNoActive bool
}

func LoadConfig(lookup func(string) (string, bool)) (Config, error) {
	userID, _ := lookup("AFDIAN_USER_ID")
	apiToken, _ := lookup("AFDIAN_API_TOKEN")

	if userID == "" || apiToken == "" {
		return Config{}, errors.New("AFDIAN_USER_ID and AFDIAN_API_TOKEN must be set")
	}

	return Config{
		UserID:   userID,
		APIToken: apiToken,
		Output: envValue(lookup, "AFDIAN_OUTPUT", "./afdian-sponsor.svg", func(value string) (string, error) {
			return value, nil
		}, func(value string) bool { return value != "" }),
		TotalSponsor:  envValue(lookup, "AFDIAN_TOTAL_SPONSORS", 100, strconv.Atoi, func(value int) bool { return value > 0 }),
		AvatarSize:    envValue(lookup, "AFDIAN_AVATAR_SIZE", 300, strconv.Atoi, func(value int) bool { return value > 0 }),
		Margin:        envValue(lookup, "AFDIAN_MARGIN", 50, strconv.Atoi, func(value int) bool { return value >= 0 }),
		AvatarsPerRow: envValue(lookup, "AFDIAN_AVATARS_PER_ROW", 15, strconv.Atoi, func(value int) bool { return value > 0 }),
		FontFile: envValue(lookup, "AFDIAN_FONT_FILE", "", func(value string) (string, error) {
			return value, nil
		}, func(value string) bool { return value != "" }),
		FontSizeScale: envValue(lookup, "AFDIAN_FONTSIZE_SCALE", 8, strconv.Atoi, func(value int) bool { return value > 0 }),
		PaddingXScale: envValue(lookup, "AFDIAN_PADDINGX_SCALE", 2, strconv.Atoi, func(value int) bool { return value >= 0 }),
		PaddingYScale: envValue(lookup, "AFDIAN_PADDINGY_SCALE", 4, strconv.Atoi, func(value int) bool { return value >= 0 }),
		Sort:          envValue(lookup, "AFDIAN_SORT", SortByTime, parseSortMode, nil),
		AnimationDelay: envValue(lookup, "AFDIAN_ANIMATION_DELAY", 0.12, parseFloat32, func(value float32) bool {
			return value >= 0
		}),
		ActiveSponsorOpacity: envValue(lookup, "AFDIAN_ACTIVE_SPONSOR_OPACITY", 1.0, parseFloat32, func(value float32) bool {
			return value >= 0 && value <= 1
		}),
		ExpiredSponsorOpacity: envValue(lookup, "AFDIAN_EXPIRED_SPONSOR_OPACITY", 0.5, parseFloat32, func(value float32) bool {
			return value >= 0 && value <= 1
		}),
		UseActiveOpacityWhenNoActive: envValue(lookup, "AFDIAN_USE_ACTIVE_OPACITY_WHEN_NO_ACTIVE", false, strconv.ParseBool, nil),
	}, nil
}

func parseFloat32(value string) (float32, error) {
	parsed, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, err
	}

	return float32(parsed), nil
}

func parseSortMode(value string) (SortMode, error) {
	mode := SortMode(strings.ToLower(value))
	switch mode {
	case SortByTime, SortByName, SortByAmount:
		return mode, nil
	default:
		return "", errors.New("sort must be time, name, or amount")
	}
}

func envValue[T any](lookup func(string) (string, bool), key string, fallback T, parse func(string) (T, error), valid func(T) bool) T {
	raw, ok := lookup(key)
	if !ok {
		return fallback
	}

	value, err := parse(raw)
	if err != nil || valid != nil && !valid(value) {
		return fallback
	}

	return value
}
