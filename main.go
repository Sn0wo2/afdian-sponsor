package main

import (
	"fmt"
	"os"
)

func main() {
	userID, apiToken, output, totalSponsor, avatarSize, margin, avatarsPerRow := getConfig()

	qs := querySponsor(userID, apiToken, totalSponsor)

	var (
		activeSponsors  []sponsor
		expiredSponsors []sponsor
	)

	for _, s := range qs {
		for _, v := range s.Data.List {
			if v.CurrentPlan.Name == "" {
				expiredSponsors = append(expiredSponsors, sponsor{
					Name:   v.User.Name,
					Avatar: v.User.Avatar,
				})
				continue
			}
			activeSponsors = append(activeSponsors, sponsor{
				Name:   v.User.Name,
				Avatar: v.User.Avatar,
			})
		}
	}

	if err := os.WriteFile(output, []byte(generateSVG(activeSponsors, expiredSponsors, avatarSize, margin, avatarsPerRow)), 0o644); err != nil { //nolint:gosec
		panic(err)
	}

	fmt.Printf("SVG file saved to %s\n", output)
}
