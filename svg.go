package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"text/template"
)

const tpl = `
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 {{.Width}} {{.Height}}">
<defs>
{{range .ActiveSponsors}}
    <clipPath id="clip-{{.Index}}"><circle cx="{{.Cx}}" cy="{{.Cy}}" r="{{.R}}"/></clipPath>
{{end}}
{{range .ExpiredSponsors}}
    <clipPath id="clip-expired-{{.Index}}"><circle cx="{{.Cx}}" cy="{{.Cy}}" r="{{.R}}"/></clipPath>
{{end}}
</defs>
<g id="active-sponsors">
{{range .ActiveSponsors}}
    <g clip-path="url(#clip-{{.Index}})">
        <title>{{.Name}}</title>
        <image x="{{.X}}" y="{{.Y}}" width="{{.AvatarSize}}" height="{{.AvatarSize}}" xlink:href="data:{{.ImgMime}};base64,{{.ImgB64}}" />
    </g>
    <text x="{{.Cx}}" y="{{.TextY}}" text-anchor="middle" font-family="Arial, sans-serif" font-size="12" fill="#333">{{.Name}}</text>
{{end}}
</g>
{{if and .ActiveSponsors .ExpiredSponsors}}
<line x1="{{.LineX1}}" y1="{{.LineY}}" x2="{{.LineX2}}" y2="{{.LineY}}" stroke="#e0e0e0" stroke-width="1"/>
{{end}}
{{if .ExpiredSponsors}}
<g id="expired-sponsors" transform="translate(0, {{.ExpiredYOffset}})">
{{range .ExpiredSponsors}}
    <g clip-path="url(#clip-expired-{{.Index}})" opacity="0.5">
        <title>{{.Name}}</title>
        <image x="{{.X}}" y="{{.Y}}" width="{{.AvatarSize}}" height="{{.AvatarSize}}" xlink:href="data:{{.ImgMime}};base64,{{.ImgB64}}" />
    </g>
    <text x="{{.Cx}}" y="{{.TextY}}" text-anchor="middle" font-family="Arial, sans-serif" font-size="12" fill="#999">{{.Name}}</text>
{{end}}
</g>
{{end}}
</svg>`

type sponsorSVG struct {
	Name       string
	Avatar     string
	Index      int
	Cx         int
	Cy         int
	X          int
	Y          int
	TextY      int
	R          int
	AvatarSize int
	ImgMime    string
	ImgB64     string
}

func generateSVG(activeSponsors, expiredSponsors []sponsorSVG, avatarSize int, margin int, avatarsPerRow int) string {
	processSponsors := func(sponsors []sponsorSVG) {
		for i := range sponsors {
			resp, err := http.Get(sponsors[i].Avatar)
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

			sponsors[i].Index = i
			sponsors[i].Cx = (i%avatarsPerRow)*(avatarSize+margin) + avatarSize/2
			sponsors[i].Cy = (i/avatarsPerRow)*(avatarSize+margin+20) + avatarSize/2
			sponsors[i].X = (i % avatarsPerRow) * (avatarSize + margin)
			sponsors[i].Y = (i / avatarsPerRow) * (avatarSize + margin + 20)
			sponsors[i].TextY = sponsors[i].Y + avatarSize + 20
			sponsors[i].R = avatarSize / 2
			sponsors[i].AvatarSize = avatarSize
			sponsors[i].ImgMime = http.DetectContentType(img)
			sponsors[i].ImgB64 = base64.StdEncoding.EncodeToString(img)
		}
	}

	processSponsors(activeSponsors)
	processSponsors(expiredSponsors)

	if len(activeSponsors) == 0 && len(expiredSponsors) == 0 {
		return `<svg width="1135" height="100" xmlns="http://www.w3.org/2000/svg" style="background-color:transparent;"></svg>`
	}

	t, err := template.New("svg").Parse(tpl)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer

	activeHeight := 0

	if len(activeSponsors) > 0 {
		activeRows := (len(activeSponsors) + avatarsPerRow - 1) / avatarsPerRow
		activeHeight = activeRows*(avatarSize+margin+40) - margin
	}

	expiredHeight := 0

	if len(expiredSponsors) > 0 {
		expiredRows := (len(expiredSponsors) + avatarsPerRow - 1) / avatarsPerRow
		expiredHeight = expiredRows*(avatarSize+margin+40) - margin
	}

	lineY := 0
	height := activeHeight + expiredHeight
	expiredYOffset := 0

	if len(activeSponsors) > 0 && len(expiredSponsors) > 0 {
		height += 40 // for separator and margin
		lineY = activeHeight + 40
		expiredYOffset = lineY + 40
	} else if len(activeSponsors) == 0 && len(expiredSponsors) > 0 {
		height = expiredHeight
	}

	if err := t.Execute(&b, struct {
		Width           int
		Height          int
		ActiveSponsors  []sponsorSVG
		ExpiredSponsors []sponsorSVG
		LineX1          int
		LineX2          int
		LineY           int
		ExpiredYOffset  int
	}{
		Width:           avatarsPerRow*(avatarSize+margin) - margin,
		Height:          height,
		ActiveSponsors:  activeSponsors,
		ExpiredSponsors: expiredSponsors,
		LineX1:          avatarSize / 2,
		LineX2:          avatarsPerRow*(avatarSize+margin) - margin - avatarSize/2,
		LineY:           lineY,
		ExpiredYOffset:  expiredYOffset,
	}); err != nil {
		panic(err)
	}

	return b.String()
}
