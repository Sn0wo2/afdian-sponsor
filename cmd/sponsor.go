package main

import (
	"cmp"
	"slices"
	"strings"
)

type Sponsor struct {
	Name        string
	Avatar      string
	TotalAmount float64
	LastPaidAt  int
}

type SponsorGroups struct {
	active  []Sponsor
	expired []Sponsor
}

func (groups *SponsorGroups) SortBy(mode SortMode) {
	var compare func(left, right Sponsor) int

	switch mode {
	case SortByName:
		compare = func(left, right Sponsor) int {
			return strings.Compare(right.Name, left.Name)
		}
	case SortByAmount:
		compare = func(left, right Sponsor) int {
			return cmp.Compare(right.TotalAmount, left.TotalAmount)
		}
	case SortByTime:
		compare = func(left, right Sponsor) int {
			return cmp.Compare(right.LastPaidAt, left.LastPaidAt)
		}
	default:
		return
	}

	slices.SortStableFunc(groups.active, compare)
	slices.SortStableFunc(groups.expired, compare)
}
