package main

import (
	"fmt"
	"os"
	"strconv"
)

func getConfig() (userID string, apiToken string, output string, page int, perPage int, avatarSize int, margin int, avatarsPerRow int) {
	userID = os.Getenv("AFDIAN_USER_ID")
	apiToken = os.Getenv("AFDIAN_API_TOKEN")

	if userID == "" || apiToken == "" {
		panic("please set AFDIAN_USER_ID and AFDIAN_API_TOKEN environment variables")
	}

	output = os.Getenv("AFDIAN_OUTPUT")
	if output == "" {
		output = "./afdian-sponsor.svg"
	}

	envPage := os.Getenv("AFDIAN_PAGE")
	envPerPage := os.Getenv("AFDIAN_PER_PAGE")
	envAvatarSize := os.Getenv("AFDIAN_AVATAR_SIZE")
	envMargin := os.Getenv("AFDIAN_MARGIN")
	envAvatarsPerRow := os.Getenv("AFDIAN_AVATARS_PER_ROW")

	page, err := strconv.Atoi(envPage)
	if err != nil || page < 1 {
		fmt.Println("AFDIAN_PAGE must be greater than 0")

		page = 1

		fmt.Println("Setting AFDIAN_PAGE to 1")
	}

	perPage, err = strconv.Atoi(envPerPage)
	if err != nil || perPage > 100 || perPage < 1 {
		fmt.Println("AFDIAN_PER_PAGE must be between 1 and 100")

		perPage = 100

		fmt.Println("Setting AFDIAN_PER_PAGE to 100")
	}

	avatarSize, err = strconv.Atoi(envAvatarSize)
	if err != nil || avatarSize < 1 {
		fmt.Println("AFDIAN_AVATAR_SIZE must be greater than 0")

		avatarSize = 100

		fmt.Println("Setting AFDIAN_AVATAR_SIZE to 100")
	}

	margin, err = strconv.Atoi(envMargin)
	if err != nil || margin < 0 {
		fmt.Println("AFDIAN_MARGIN must be greater than or equal to 0")

		margin = 15

		fmt.Println("Setting AFDIAN_MARGIN to 15")
	}

	avatarsPerRow, err = strconv.Atoi(envAvatarsPerRow)
	if err != nil || avatarsPerRow < 1 {
		fmt.Println("AFDIAN_AVATARS_PER_ROW must be greater than 0")

		avatarsPerRow = 10

		fmt.Println("Setting AFDIAN_AVATARS_PER_ROW to 10")
	}

	return
}
