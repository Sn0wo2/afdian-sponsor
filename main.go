package main

import (
	"fmt"
	"os"
)

func main() {
	userID, apiToken, output, page, perPage, avatarSize, margin, avatarsPerRow := getConfig()

	sponsor := querySponsor(userID, apiToken, page, perPage)

	err := os.WriteFile(output, []byte(generateSVG(sponsor, avatarSize, margin, avatarsPerRow)), 0o600)
	if err != nil {
		panic(err)
	}

	fmt.Printf("SVG file saved to %s\n", output)
}
