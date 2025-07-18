package main

import (
	"fmt"
	"os"
)

func main() {
	userID, apiToken, output, page, perPage, avatarSize, margin, avatarsPerRow := getConfig()

	if err := os.WriteFile(output, []byte(generateSVG(querySponsor(userID, apiToken, page, perPage), avatarSize, margin, avatarsPerRow)), 0o644); err != nil {
		panic(err)
	}

	fmt.Printf("SVG file saved to %s\n", output)
}
