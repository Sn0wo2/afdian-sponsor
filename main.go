package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/Sn0wo2/afdian-sponsor/version"
)

func main() {
	fmt.Printf("afdian-sponsor %s-%s(%s)\n", version.GetVersion(), version.GetCommit(), version.GetDate())

	cfg := GetConfig()

	qs := QuerySponsor(cfg.UserID, cfg.APIToken, cfg.TotalSponsor)

	var (
		activeSponsors  []Sponsor
		expiredSponsors []Sponsor
	)

	for _, s := range qs {
		for _, v := range s.Data.List {
			amount, err := strconv.ParseFloat(v.AllSumAmount, 64)
			if err != nil {
				fmt.Printf("Warning: Failed to parse AllSumAmount for sponsor %s: %v\n", v.User.Name, err)

				amount = 0
			}

			if v.CurrentPlan.Name == "" {
				expiredSponsors = append(expiredSponsors, Sponsor{
					Name:         v.User.Name,
					Avatar:       v.User.Avatar,
					AllSumAmount: amount,
					LastPayTime:  v.LastPayTime,
				})

				continue
			}

			activeSponsors = append(activeSponsors, Sponsor{
				Name:         v.User.Name,
				Avatar:       v.User.Avatar,
				AllSumAmount: amount,
				LastPayTime:  v.LastPayTime,
			})
		}
	}

	switch cfg.Sort {
	case "name":
		sort.Slice(activeSponsors, func(i, j int) bool {
			return activeSponsors[i].Name > activeSponsors[j].Name
		})
		sort.Slice(expiredSponsors, func(i, j int) bool {
			return expiredSponsors[i].Name > expiredSponsors[j].Name
		})
	case "amount":
		sort.Slice(activeSponsors, func(i, j int) bool {
			return activeSponsors[i].AllSumAmount > activeSponsors[j].AllSumAmount
		})
		sort.Slice(expiredSponsors, func(i, j int) bool {
			return expiredSponsors[i].AllSumAmount > expiredSponsors[j].AllSumAmount
		})
	// time
	default:
		sort.Slice(activeSponsors, func(i, j int) bool {
			return activeSponsors[i].LastPayTime > activeSponsors[j].LastPayTime
		})
		sort.Slice(expiredSponsors, func(i, j int) bool {
			return expiredSponsors[i].LastPayTime > expiredSponsors[j].LastPayTime
		})
	}

	if err := os.WriteFile(cfg.Output, []byte(Generate(activeSponsors, expiredSponsors, cfg.AvatarSize, cfg.Margin, cfg.AvatarsPerRow, cfg.AnimationDelay)), 0o644); err != nil { //nolint:gosec
		panic(err)
	}

	fmt.Printf("SVG file saved to %s\n", cfg.Output)
}
