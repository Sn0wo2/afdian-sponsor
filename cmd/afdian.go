package main

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/Sn0wo2/go-afdian-api/v2"
	"github.com/Sn0wo2/go-afdian-api/v2/pkg/payload"
)

func QuerySponsors(client *http.Client, userID, apiToken string, limit int) (SponsorGroups, error) {
	if limit <= 0 {
		return SponsorGroups{}, fmt.Errorf("sponsor limit must be positive: %d", limit)
	}

	api := afdian.NewClient(&afdian.Config{
		UserID:   userID,
		APIToken: apiToken,
	}, client)
	perPage := min(limit, 100)

	firstPage, err := api.QuerySponsor(1, perPage)
	if err != nil {
		return SponsorGroups{}, fmt.Errorf("query sponsor page 1: %w", err)
	}

	var groups SponsorGroups

	appendPage := func(page *payload.QuerySponsor, remaining int) int {
		entries := page.Data.List[:min(len(page.Data.List), remaining)]

		for _, entry := range entries {
			if entry.User == nil {
				_, _ = fmt.Fprintln(os.Stderr, "warning: skipping sponsor without user data")

				continue
			}

			amount, err := strconv.ParseFloat(entry.AllSumAmount, 64)
			if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) {
				_, _ = fmt.Fprintf(os.Stderr, "warning: invalid total amount %q for sponsor %q\n", entry.AllSumAmount, entry.User.Name)

				amount = 0
			}

			item := Sponsor{
				Name:        entry.User.Name,
				Avatar:      entry.User.Avatar,
				TotalAmount: amount,
				LastPaidAt:  entry.LastPayTime,
			}
			if entry.CurrentPlan == nil || entry.CurrentPlan.Name == "" {
				groups.expired = append(groups.expired, item)
			} else {
				groups.active = append(groups.active, item)
			}
		}

		return len(entries)
	}

	total := min(limit, firstPage.Data.TotalCount)
	if total <= 0 {
		appendPage(firstPage, limit)

		return groups, nil
	}

	pageCount := max(min((total+perPage-1)/perPage, firstPage.Data.TotalPage), 1)
	seen := appendPage(firstPage, total)

	for pageNumber := 2; pageNumber <= pageCount && seen < total; pageNumber++ {
		page, err := api.QuerySponsor(pageNumber, perPage)
		if err != nil {
			return SponsorGroups{}, fmt.Errorf("query sponsor page %d: %w", pageNumber, err)
		}

		seen += appendPage(page, total-seen)
	}

	return groups, nil
}
