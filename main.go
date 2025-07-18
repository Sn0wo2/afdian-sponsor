package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Sn0wo2/go-afdian-api"
)

func main() {
	userID := os.Getenv("AFDIAN_USER_ID")
	apiToken := os.Getenv("AFDIAN_API_TOKEN")

	if userID == "" || apiToken == "" {
		panic("please set AFDIAN_USER_ID and AFDIAN_API_TOKEN environment variables")
	}

	output := os.Getenv("AFDIAN_OUTPUT")
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
	}

	perPage, err := strconv.Atoi(envPerPage)
	if err != nil || perPage > 100 || perPage < 1 {
		fmt.Println("AFDIAN_PER_PAGE must be between 1 and 100")

		perPage = 100
	}

	avatarSize, err := strconv.Atoi(envAvatarSize)
	if err != nil || avatarSize < 1 {
		fmt.Println("AFDIAN_AVATAR_SIZE must be greater than 0")

		avatarSize = 100
	}

	margin, err := strconv.Atoi(envMargin)
	if err != nil || margin < 0 {
		fmt.Println("AFDIAN_MARGIN must be greater than or equal to 0")

		margin = 20
	}

	avatarsPerRow, err := strconv.Atoi(envAvatarsPerRow)
	if err != nil || avatarsPerRow < 1 {
		fmt.Println("AFDIAN_AVATARS_PER_ROW must be greater than 0")

		avatarsPerRow = 6
	}

	config := &afdian.Config{
		UserID:   userID,
		APIToken: apiToken,
	}

	client := afdian.NewClient(config)

	sponsor, err := client.QuerySponsor(page, perPage)
	if err != nil {
		panic(err)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 %d %d">`, avatarsPerRow*(avatarSize+margin)-margin, (len(sponsor.Data.List)+avatarsPerRow-1)/avatarsPerRow*(avatarSize+margin)-margin))
	sb.WriteString(`<defs>`)

	for i := range sponsor.Data.List {
		sb.WriteString(fmt.Sprintf(`<clipPath id="clip-%d"><circle cx="%d" cy="%d" r="%d"/></clipPath>`, i, (i%avatarsPerRow)*(avatarSize+margin)+avatarSize/2, (i/avatarsPerRow)*(avatarSize+margin)+avatarSize/2, avatarSize/2))
	}

	sb.WriteString("</defs>\n")

	for i, v := range sponsor.Data.List {
		x := (i % avatarsPerRow) * (avatarSize + margin)
		y := (i / avatarsPerRow) * (avatarSize + margin)

		resp, err := http.Get(v.User.Avatar)
		if err != nil {
			panic(err)
		}

		defer func() 
		defer func() {
			_ = resp.Body.Close()
		}()

		img, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		sb.WriteString(fmt.Sprintf(`<g clip-path="url(#clip-%d)">`, i))
		sb.WriteString(fmt.Sprintf(`<title>%s</title>`, v.User.Name))
		sb.WriteString(fmt.Sprintf(`<image x="%d" y="%d" width="%d" height="%d" xlink:href="data:%s;base64,%s" />`, x, y, avatarSize, avatarSize, http.DetectContentType(img), base64.StdEncoding.EncodeToString(img)))
		sb.WriteString(`</g>` + "\n")
	}

	sb.WriteString(`</svg>`)

	err = os.WriteFile(output, []byte(sb.String()), 0o600)
	if err != nil {
		panic(err)
	}

	fmt.Printf("SVG file saved to %s", output)
}
