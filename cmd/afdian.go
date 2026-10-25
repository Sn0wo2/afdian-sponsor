package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/Sn0wo2/go-afdian-api"
	"github.com/Sn0wo2/go-afdian-api/pkg/payload"
)

const MaxSponsorsPerPage = 100

func QuerySponsors(client *http.Client, userID, apiToken string, limit int) (SponsorGroups, error) {
	if limit <= 0 {
		return SponsorGroups{}, fmt.Errorf("sponsor limit must be positive: %d", limit)
	}

	api := afdian.NewClient(&afdian.Config{
		UserID:   userID,
		APIToken: apiToken,
	}, client)
	perPage := min(limit, MaxSponsorsPerPage)

	firstPage, err := api.QuerySponsor(1, perPage)
	if err != nil {
		return SponsorGroups{}, fmt.Errorf("query sponsor page 1: %w", err)
	}

	if firstPage == nil {
		return SponsorGroups{}, errors.New("query sponsor page 1: empty response")
	}

	var groups SponsorGroups

	appendPage := func(page *payload.QuerySponsor, remaining int) int {
		seen := 0
		for _, entry := range page.Data.List {
			if seen == remaining {
				break
			}

			seen++

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

		return seen
	}

	total := min(limit, firstPage.Data.TotalCount)
	if total <= 0 {
		appendPage(firstPage, limit)

		return groups, nil
	}

	pageCount := min((total+perPage-1)/perPage, firstPage.Data.TotalPage)
	pageCount = max(pageCount, 1)
	seen := appendPage(firstPage, total)

	for pageNumber := 2; pageNumber <= pageCount && seen < total; pageNumber++ {
		page, err := api.QuerySponsor(pageNumber, perPage)
		if err != nil {
			return SponsorGroups{}, fmt.Errorf("query sponsor page %d: %w", pageNumber, err)
		}

		if page == nil {
			return SponsorGroups{}, fmt.Errorf("query sponsor page %d: empty response", pageNumber)
		}

		seen += appendPage(page, total-seen)
	}

	return groups, nil
}
