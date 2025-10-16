package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"text/template"

	common "github.com/Sn0wo2/afdian-sponsor/internal/helper"
)

const svgTPL = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 {{.Width}} {{.Height}}">
<style>
    .active-text { fill: #000000; font-weight: bold; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; }
    .expired-text { fill: #666666; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; }
    .separator { stroke: #eeeeee; }
    @media (prefers-color-scheme: dark) {
        .active-text, .expired-text { fill: #fff; }
        .separator { stroke: #333; }
    }
    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateX(-20px);
        }
        to {
            opacity: 1;
            transform: translateX(0);
        }
    }
</style>
<defs>
{{range .ActiveSponsors}}
    <clipPath id="clip-{{.Index}}"><circle cx="{{.CenterX}}" cy="{{.CenterY}}" r="{{.Radius}}"/></clipPath>
{{end}}
{{range .ExpiredSponsors}}
    <clipPath id="clip-expired-{{.Index}}"><circle cx="{{.CenterX}}" cy="{{.CenterY}}" r="{{.Radius}}"/></clipPath>
{{end}}
</defs>
<g id="active-sponsors">
{{range .ActiveSponsors}}
    <g style="animation: fadeIn 0.5s ease-in-out forwards; animation-delay: {{.AnimationDelay}}s; opacity: 0;">
        <g clip-path="url(#clip-{{.Index}})">
            <title>{{.OriginalName}}</title>
            <image x="{{.X}}" y="{{.Y}}" width="{{.AvatarSize}}" height="{{.AvatarSize}}" xlink:href="data:{{.ImgMime}};base64,{{.ImgB64}}" />
        </g>
        <text x="{{.CenterX}}" y="{{.TextY}}" text-anchor="middle" font-size="{{$.FontSize}}" class="active-text">{{.Name}}</text>
    </g>
{{end}}
</g>
{{if and .ActiveSponsors .ExpiredSponsors}}
<line class="separator" x1="{{.LineX1}}" y1="{{.LineY}}" x2="{{.LineX2}}" y2="{{.LineY}}" stroke-width="1"/>
{{end}}
{{if .ExpiredSponsors}}
<g id="expired-sponsors">
{{range .ExpiredSponsors}}
    <g style="animation: fadeIn 0.5s ease-in-out forwards; animation-delay: {{.AnimationDelay}}s; opacity: 0;">
        <g clip-path="url(#clip-expired-{{.Index}})" opacity="0.5">
            <title>{{.OriginalName}}</title>
            <image x="{{.X}}" y="{{.Y}}" width="{{.AvatarSize}}" height="{{.AvatarSize}}" xlink:href="data:{{.ImgMime}};base64,{{.ImgB64}}" />
        </g>
        <text x="{{.CenterX}}" y="{{.TextY}}" text-anchor="middle" font-size="{{$.FontSize}}" class="expired-text">{{.Name}}</text>
    </g>
{{end}}
</g>
{{end}}
</svg>`

const emptySVG = `<svg width="1135" height="100" xmlns="http://www.w3.org/2000/svg" style="background-color:transparent;"></svg>`

// Generate generates an SVG from the given sponsors.
func Generate(activeSponsors, expiredSponsors []Sponsor, avatarSize int, margin int, avatarsPerRow int, animationDelay float32) (string, error) {
	if len(activeSponsors) == 0 && len(expiredSponsors) == 0 {
		return emptySVG, nil
	}

	fontSize := avatarSize / 8

	nameLimit := avatarSize * 2 / fontSize
	if nameLimit < 5 {
		nameLimit = 5
	}

	paddingX := avatarSize / 2
	paddingY := avatarSize / 4
	rowHeight := avatarSize + margin + fontSize + 10
	textYMargin := avatarSize + fontSize + 10

	processSponsors := func(sponsors []Sponsor, startY int, active ...bool) error {
		for i := range sponsors {
			sponsors[i].OriginalName = sponsors[i].Name
			if common.StringWidth(sponsors[i].Name) > nameLimit {
				sponsors[i].Name = common.TruncateStringByWidth(sponsors[i].Name, nameLimit)
			}

			resp, err := http.Get(sponsors[i].Avatar)
			if err != nil {
				return err
			}

			img, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			_ = resp.Body.Close()

			sponsors[i].Index = i
			sponsors[i].Y = startY + (i/avatarsPerRow)*rowHeight
			sponsors[i].X = paddingX + (i%avatarsPerRow)*(avatarSize+margin)
			sponsors[i].CenterY = sponsors[i].Y + avatarSize/2
			sponsors[i].CenterX = sponsors[i].X + avatarSize/2
			sponsors[i].TextY = sponsors[i].Y + textYMargin
			sponsors[i].Radius = avatarSize / 2
			sponsors[i].AvatarSize = avatarSize
			sponsors[i].ImgMime = http.DetectContentType(img)
			sponsors[i].ImgB64 = base64.StdEncoding.EncodeToString(img)

			animationIndex := float32(i)
			if len(active) > 0 && active[0] {
				sponsors[i].AnimationDelay = animationIndex * animationDelay
			} else {
				sponsors[i].AnimationDelay = animationIndex * animationDelay / 1.5
			}
		}
		return nil
	}

	if err := processSponsors(activeSponsors, paddingY, true); err != nil {
		return emptySVG, err
	}

	numActiveRows := (len(activeSponsors) + avatarsPerRow - 1) / avatarsPerRow
	activeHeight := numActiveRows * rowHeight

	separatorHeight := 0
	if len(activeSponsors) > 0 && len(expiredSponsors) > 0 {
		separatorHeight = 40
	}

	expiredStartY := paddingY + activeHeight + separatorHeight
	err := processSponsors(expiredSponsors, expiredStartY)
	if err != nil {
		return emptySVG, err
	}

	numExpiredRows := (len(expiredSponsors) + avatarsPerRow - 1) / avatarsPerRow
	expiredHeight := numExpiredRows * rowHeight

	height := paddingY + activeHeight + separatorHeight + expiredHeight + paddingY

	contentWidth := 0
	if len(activeSponsors) > 0 || len(expiredSponsors) > 0 {
		contentWidth = avatarsPerRow*(avatarSize+margin) - margin
	}

	width := contentWidth + paddingX*2

	t, err := template.New("svg").Parse(svgTPL)
	if err != nil {
		return emptySVG, err
	}

	var b bytes.Buffer

	err = t.Execute(&b, struct {
		Width           int
		Height          int
		FontSize        int
		ActiveSponsors  []Sponsor
		ExpiredSponsors []Sponsor
		LineX1          int
		LineX2          int
		LineY           int
		ExpiredYOffset  int
	}{
		Width:           width,
		Height:          height,
		FontSize:        fontSize,
		ActiveSponsors:  activeSponsors,
		ExpiredSponsors: expiredSponsors,
		LineX1:          paddingX,
		LineX2:          width - paddingX,
		LineY:           paddingY + activeHeight + separatorHeight/2,
		ExpiredYOffset:  0,
	})

	return b.String(), err
}
