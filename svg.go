package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Sn0wo2/go-afdian-api/pkg/payload"
)

func generateSVG(sponsor *payload.QuerySponsor, avatarSize int, margin int, avatarsPerRow int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 %d %d">`, avatarsPerRow*(avatarSize+margin)-margin, (len(sponsor.Data.List)+avatarsPerRow-1)/avatarsPerRow*(avatarSize+margin)-margin))
	sb.WriteString(`<defs>`)

	for i := range sponsor.Data.List {
		sb.WriteString(fmt.Sprintf(`<clipPath id="clip-%d"><circle cx="%d" cy="%d" r="%d"/></clipPath>`, i, (i%avatarsPerRow)*(avatarSize+margin)+avatarSize/2, (i/avatarsPerRow)*(avatarSize+margin)+avatarSize/2, avatarSize/2))
	}

	sb.WriteString("</defs>\n")

	for i, v := range sponsor.Data.List {
		x := (i % avatarsPerRow) * (avatarSize + margin)
		y := (i / avatarsPerRow) * (avatarSize + margin)

		resp, err := http.Get(v.User.Avatar)
		if err != nil {
			panic(err)
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		img, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		sb.WriteString(fmt.Sprintf(`<g clip-path="url(#clip-%d)">`, i))
		sb.WriteString(fmt.Sprintf(`<title>%s</title>`, v.User.Name))
		sb.WriteString(fmt.Sprintf(`<image x="%d" y="%d" width="%d" height="%d" xlink:href="data:%s;base64,%s" />`, x, y, avatarSize, avatarSize, http.DetectContentType(img), base64.StdEncoding.EncodeToString(img)))
		sb.WriteString(`</g>` + "\n")
	}

	sb.WriteString(`</svg>`)

	return sb.String()
}
