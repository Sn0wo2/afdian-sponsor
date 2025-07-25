package main

import (
	"fmt"
	"os"

	"github.com/Sn0wo2/afdian-sponsor/internal/afdian"
	"github.com/Sn0wo2/afdian-sponsor/internal/config"
	"github.com/Sn0wo2/afdian-sponsor/internal/svg"
	"github.com/Sn0wo2/afdian-sponsor/internal/types"
	"github.com/Sn0wo2/afdian-sponsor/internal/version"
)

func main() {
	fmt.Printf("%s-%s(%s)\n", version.GetVersion(), version.GetCommit(), version.GetDate())

	cfg := config.GetConfig()

	qs := afdian.QuerySponsor(cfg.UserID, cfg.APIToken, cfg.TotalSponsor)

	var (
		activeSponsors  []types.Sponsor
		expiredSponsors []types.Sponsor
	)

	for _, s := range qs {
		for _, v := range s.Data.List {
			if v.CurrentPlan.Name == "" {
				expiredSponsors = append(expiredSponsors, types.Sponsor{
					Name:   v.User.Name,
					Avatar: v.User.Avatar,
				})

				continue
			}

			activeSponsors = append(activeSponsors, types.Sponsor{
				Name:   v.User.Name,
				Avatar: v.User.Avatar,
			})
		}
	}

	if err := os.WriteFile(cfg.Output, []byte(svg.Generate(activeSponsors, expiredSponsors, cfg.AvatarSize, cfg.Margin, cfg.AvatarsPerRow)), 0o644); err != nil { //nolint:gosec
		panic(err)
	}

	fmt.Printf("SVG file saved to %s\n", cfg.Output)
}
