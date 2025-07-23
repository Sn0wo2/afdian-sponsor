package main

import (
	"github.com/Sn0wo2/go-afdian-api"
	"github.com/Sn0wo2/go-afdian-api/pkg/payload"
)

func querySponsor(userID string, apiToken string, totalSponsor int) []*payload.QuerySponsor {
	page := 1
	perPage := 100
	if totalSponsor < 100 {
		perPage = totalSponsor
	}

	sponsor, err := afdian.NewClient(&afdian.Config{
		UserID:   userID,
		APIToken: apiToken,
	}).QuerySponsor(page, perPage)
	if err != nil {
		panic(err)
	}

	if totalSponsor > sponsor.Data.TotalCount {
		totalSponsor = sponsor.Data.TotalCount
	}

	fetchPage := (totalSponsor + perPage - 1) / perPage

	if fetchPage > sponsor.Data.TotalPage {
		fetchPage = sponsor.Data.TotalPage
	}

	sponsors := make([]*payload.QuerySponsor, 0, fetchPage)
	sponsors = append(sponsors, sponsor)

	if fetchPage <= sponsor.Data.TotalPage {
		return sponsors
	}

	for i := 2; i <= fetchPage; i++ {
		sponsor, err = afdian.NewClient(&afdian.Config{
			UserID:   userID,
			APIToken: apiToken,
		}).QuerySponsor(i, perPage)
		if err != nil {
			panic(err)
		}
		sponsors = append(sponsors, sponsor)
	}

	return sponsors
}
