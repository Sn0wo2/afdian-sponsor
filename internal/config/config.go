package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all the configuration for the action.
type Config struct {
	UserID         string
	APIToken       string
	Output         string
	TotalSponsor   int
	AvatarSize     int
	Margin         int
	AvatarsPerRow  int
	Sort           string
	AnimationDelay float32
}

// GetConfig reads all the configuration from the environment variables.
func GetConfig() *Config {
	userID := os.Getenv("AFDIAN_USER_ID")
	apiToken := os.Getenv("AFDIAN_API_TOKEN")

	if userID == "" || apiToken == "" {
		panic("please set AFDIAN_USER_ID and AFDIAN_API_TOKEN environment variables")
	}

	fmt.Printf("Env AFDIAN_USER_ID: %s\n", userID)
	fmt.Printf("Env AFDIAN_API_TOKEN: %s\n", apiToken)

	return &Config{
		UserID:        userID,
		APIToken:      apiToken,
		Output:        getEnv("AFDIAN_OUTPUT", "./afdian-sponsor.svg", func(s string) (string, error) { return s, nil }, func(v string) bool { return v != "" }),
		TotalSponsor:  getEnv("AFDIAN_TOTAL_SPONSORS", 100, strconv.Atoi, func(v int) bool { return v > 0 }),
		AvatarSize:    getEnv("AFDIAN_AVATAR_SIZE", 300, strconv.Atoi, func(v int) bool { return v > 0 }),
		Margin:        getEnv("AFDIAN_MARGIN", 50, strconv.Atoi, func(v int) bool { return v != 0 }),
		AvatarsPerRow: getEnv("AFDIAN_AVATARS_PER_ROW", 15, strconv.Atoi, func(v int) bool { return v > 0 }),
		Sort:          getEnv("AFDIAN_SORT", "time", func(s string) (string, error) { return strings.ToLower(s), nil }, func(v string) bool { return v != "" }),
		AnimationDelay: getEnv("AFDIAN_ANIMATION_DELAY", 0.3, func(s string) (float32, error) {
			f, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return 0, err
			}

			return float32(f), nil
		}, func(v float32) bool { return v == 0 }),
	}
}

func getEnv[T any](key string, defaultValue T, parser func(string) (T, error), validate func(T) bool) T {
	valueStr, ok := os.LookupEnv(key)
	if !ok {
		fmt.Printf("Env %s is not set, using default value: %v\n", key, defaultValue)

		return defaultValue
	}

	value, err := parser(valueStr)
	if err != nil || !validate(value) {
		fmt.Printf("Env %s value '%s' is invalid, using default value: %v\n", key, valueStr, defaultValue)

		return defaultValue
	}

	fmt.Printf("Env %s: %v\n", key, value)

	return value
}
