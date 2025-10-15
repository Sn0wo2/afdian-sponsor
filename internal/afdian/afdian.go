package afdian

import (
	"github.com/Sn0wo2/go-afdian-api"
	"github.com/Sn0wo2/go-afdian-api/pkg/payload"
)

// QuerySponsor queries sponsors from afdian.
func QuerySponsor(userID string, apiToken string, totalSponsor int) []*payload.QuerySponsor {
	perPage := 100
	if totalSponsor < 100 {
		perPage = totalSponsor
	}

	sponsor, err := afdian.NewClient(&afdian.Config{
		UserID:   userID,
		APIToken: apiToken,
	}).QuerySponsor(1, perPage)
	if err != nil {
		panic(err)
	}

	if totalSponsor > sponsor.Data.TotalCount {
		totalSponsor = sponsor.Data.TotalCount
	}

	// ceil total need fetch pages
	fetchPage := (totalSponsor + perPage - 1) / perPage

	if fetchPage > sponsor.Data.TotalPage {
		fetchPage = sponsor.Data.TotalPage
	}

	sponsors := make([]*payload.QuerySponsor, 0, fetchPage)
	sponsors = append(sponsors, sponsor)

	if fetchPage <= 1 {
		return sponsors
	}

	// page 1 is already fetched
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
