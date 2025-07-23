package main

import (
	"fmt"
	"os"
	"strconv"
)

func getConfig() (userID string, apiToken string, output string, totalSponsor int, avatarSize int, margin int, avatarsPerRow int) {
	userID = os.Getenv("AFDIAN_USER_ID")
	apiToken = os.Getenv("AFDIAN_API_TOKEN")

	if userID == "" || apiToken == "" {
		panic("please set AFDIAN_USER_ID and AFDIAN_API_TOKEN environment variables")
	}

	fmt.Printf("Env AFDIAN_USER_ID: %s\n", userID)
	fmt.Printf("Env AFDIAN_API_TOKEN: %s\n", apiToken)

	parseString := func(s string) (string, error) { return s, nil }
	parseInt := func(s string) (int, error) { return strconv.Atoi(s) }

	output = getEnv("AFDIAN_OUTPUT", "./afdian-sponsor.svg", parseString, func(v string) bool { return v != "" })
	totalSponsor = getEnv("AFDIAN_TOTAL_SPONSORS", 100, parseInt, func(v int) bool { return v > 0 })
	avatarSize = getEnv("AFDIAN_AVATAR_SIZE", 100, parseInt, func(v int) bool { return v > 0 })
	margin = getEnv("AFDIAN_MARGIN", 15, parseInt, func(v int) bool { return v != 0 })
	avatarsPerRow = getEnv("AFDIAN_AVATARS_PER_ROW", 10, parseInt, func(v int) bool { return v > 0 })

	return
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
