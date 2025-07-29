package svg

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"text/template"

	"github.com/Sn0wo2/afdian-sponsor/internal/types"
	"github.com/mattn/go-runewidth"
)

const tpl = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 {{.Width}} {{.Height}}">
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
        <text x="{{.CenterX}}" y="{{.TextY}}" text-anchor="middle" font-size="12" class="active-text">{{.Name}}</text>
    </g>
{{end}}
</g>
{{if and .ActiveSponsors .ExpiredSponsors}}
<line class="separator" x1="{{.LineX1}}" y1="{{.LineY}}" x2="{{.LineX2}}" y2="{{.LineY}}" stroke-width="1"/>
{{end}}
{{if .ExpiredSponsors}}
<g id="expired-sponsors" transform="translate(0, {{.ExpiredYOffset}})">
{{range .ExpiredSponsors}}
    <g style="animation: fadeIn 0.5s ease-in-out forwards; animation-delay: {{.AnimationDelay}}s; opacity: 0;">
        <g clip-path="url(#clip-expired-{{.Index}})" opacity="0.5">
            <title>{{.OriginalName}}</title>
            <image x="{{.X}}" y="{{.Y}}" width="{{.AvatarSize}}" height="{{.AvatarSize}}" xlink:href="data:{{.ImgMime}};base64,{{.ImgB64}}" />
        </g>
        <text x="{{.CenterX}}" y="{{.TextY}}" text-anchor="middle" font-size="12" class="expired-text">{{.Name}}</text>
    </g>
{{end}}
</g>
{{end}}
</svg>`

func stringWidth(s string) int {
	return runewidth.StringWidth(s)
}

func truncateStringByWidth(s string, limit int) string {
	if stringWidth(s) <= limit {
		return s
	}

	width := 0

	runes := []rune(s)
	for i, r := range runes {
		width += runewidth.RuneWidth(r)
		if width > limit {
			return string(runes[:i]) + "..."
		}
	}

	return s
}

// Generate generates an SVG from the given sponsors.
func Generate(activeSponsors, expiredSponsors []types.Sponsor, avatarSize int, margin int, avatarsPerRow int, animationDelay float32) string {
	nameLimit := avatarSize / 6
	if nameLimit < 5 {
		nameLimit = 5
	}

	processSponsors := func(sponsors []types.Sponsor, active ...bool) {
		rowHeight := avatarSize + margin + 35
		textYMargin := avatarSize + 25

		for i := range sponsors {
			sponsors[i].OriginalName = sponsors[i].Name
			if stringWidth(sponsors[i].Name) > nameLimit {
				sponsors[i].Name = truncateStringByWidth(sponsors[i].Name, nameLimit)
			}

			resp, err := http.Get(sponsors[i].Avatar)
			if err != nil {
				panic(err)
			}

			img, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			_ = resp.Body.Close()

			sponsors[i].Index = i
			sponsors[i].CenterX = (i%avatarsPerRow)*(avatarSize+margin) + avatarSize/2
			sponsors[i].CenterY = (i/avatarsPerRow)*rowHeight + avatarSize/2
			sponsors[i].X = (i % avatarsPerRow) * (avatarSize + margin)
			sponsors[i].Y = (i / avatarsPerRow) * rowHeight
			sponsors[i].TextY = sponsors[i].Y + textYMargin
			sponsors[i].Radius = avatarSize / 2
			sponsors[i].AvatarSize = avatarSize
			sponsors[i].ImgMime = http.DetectContentType(img)
			sponsors[i].ImgB64 = base64.StdEncoding.EncodeToString(img)
			if len(active) > 0 && active[0] {
				sponsors[i].AnimationDelay = float32(i) * animationDelay
			} else {
				sponsors[i].AnimationDelay = float32(i) * animationDelay / 1.5
			}
		}
	}

	processSponsors(activeSponsors, true)
	processSponsors(expiredSponsors)

	if len(activeSponsors) == 0 && len(expiredSponsors) == 0 {
		return `<svg width="1135" height="100" xmlns="http://www.w3.org/2000/svg" style="background-color:transparent;"></svg>`
	}

	t, err := template.New("svg").Parse(tpl)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer

	rowHeight := avatarSize + margin + 35
	separatorHeight := 40

	calculateHeight := func(sponsors []types.Sponsor) int {
		if len(sponsors) == 0 {
			return 0
		}

		rows := (len(sponsors) + avatarsPerRow - 1) / avatarsPerRow

		return rows*rowHeight - margin
	}

	activeHeight := calculateHeight(activeSponsors)
	expiredHeight := calculateHeight(expiredSponsors)

	height := activeHeight + expiredHeight
	if activeHeight > 0 && expiredHeight > 0 {
		height += separatorHeight
	}

	lineY := activeHeight + separatorHeight/2

	expiredYOffset := activeHeight
	if activeHeight > 0 && expiredHeight > 0 {
		expiredYOffset += separatorHeight
	}

	if err := t.Execute(&b, struct {
		Width           int
		Height          int
		ActiveSponsors  []types.Sponsor
		ExpiredSponsors []types.Sponsor
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
