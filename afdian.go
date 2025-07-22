package main

import (
	"github.com/Sn0wo2/go-afdian-api"
	"github.com/Sn0wo2/go-afdian-api/pkg/payload"
)

func querySponsor(userID string, apiToken string, page int, perPage int) *payload.QuerySponsor {
	sponsor, err := afdian.NewClient(&afdian.Config{
		UserID:   userID,
		APIToken: apiToken,
	}).QuerySponsor(page, perPage)

	if err != nil {
		panic(err)
	}

	return sponsor
}
