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
    @keyframes pop-out {
        from {
            opacity: 0;
            transform: translate(var(--tx, 0), var(--ty, 0)) scale(0.5);
        }
        to {
            opacity: 1;
            transform: translate(0, 0) scale(1);
        }
    }
    .sponsor-group {
        animation: pop-out 0.5s cubic-bezier(0.25, 1, 0.5, 1) forwards;
        opacity: 0;
    }
</style>
<defs>
{{range .Sponsors}}
    <clipPath id="clip-{{.Index}}"><circle cx="0" cy="0" r="{{.Radius}}"/></clipPath>
{{end}}
</defs>
{{range .Sponsors}}
    <g transform="translate({{.CenterX}}, {{.CenterY}})" style="opacity: {{.Opacity}};">
        <g class="sponsor-group" style="animation-delay: {{.AnimationDelay}}s; --tx: {{.TranslateX}}px; --ty: {{.TranslateY}}px;">
            <g clip-path="url(#clip-{{.Index}})">
                <title>{{.OriginalName}}</title>
                <image x="-{{.Radius}}" y="-{{.Radius}}" width="{{.AvatarSize}}" height="{{.AvatarSize}}" xlink:href="data:{{.ImgMime}};base64,{{.ImgB64}}" />
            </g>
            <text y="{{.TextY}}" text-anchor="middle" font-size="{{$.FontSize}}" class="{{if .IsActive}}active-text{{else}}expired-text{{end}}">{{.Name}}</text>
        </g>
    </g>
{{end}}
{{if and .ActiveSponsors .ExpiredSponsors}}
<line class="separator" x1="{{.LineX1}}" y1="{{.LineY}}" x2="{{.LineX2}}" y2="{{.LineY}}" stroke-width="1"/>
{{end}}
</svg>`

const emptySVG = `<svg width="1135" height="100" xmlns="http://www.w3.org/2000/svg" style="background-color:transparent;"></svg>`

// Generate generates an SVG from the given sponsors.
func Generate(client *http.Client, activeSponsors, expiredSponsors []Sponsor, cfg *Config) (string, error) {
	if len(activeSponsors) == 0 && len(expiredSponsors) == 0 {
		return emptySVG, nil
	}

	fontSize := cfg.AvatarSize / cfg.FontSizeScale

	nameLimit := cfg.AvatarSize * 2 / fontSize
	if nameLimit < 5 {
		nameLimit = 5
	}

	paddingX := 0
	if cfg.PaddingXScale > 0 {
		paddingX = cfg.AvatarSize / cfg.PaddingXScale
	}

	paddingY := 0
	if cfg.PaddingYScale > 0 {
		paddingY = cfg.AvatarSize / cfg.PaddingYScale
	}

	rowHeight := cfg.AvatarSize + cfg.Margin + fontSize + 10
	textYOffset := cfg.AvatarSize/2 + fontSize + 10

	numActiveRows := (len(activeSponsors) + cfg.AvatarsPerRow - 1) / cfg.AvatarsPerRow
	activeHeight := numActiveRows * rowHeight

	separatorHeight := 0
	if len(activeSponsors) > 0 && len(expiredSponsors) > 0 {
		separatorHeight = 40
	}

	numExpiredRows := (len(expiredSponsors) + cfg.AvatarsPerRow - 1) / cfg.AvatarsPerRow
	expiredHeight := numExpiredRows * rowHeight
	height := paddingY + activeHeight + separatorHeight + expiredHeight
	width := cfg.AvatarsPerRow*(cfg.AvatarSize+cfg.Margin) - cfg.Margin + paddingX*2
	svgCenterX := width / 2
	svgCenterY := height / 2

	var allSponsors []Sponsor

	allSponsors = append(allSponsors, activeSponsors...)
	allSponsors = append(allSponsors, expiredSponsors...)

	processSponsor := func(sponsor *Sponsor, index int, isActive bool, activeCount int) error {
		sponsor.OriginalName = sponsor.Name
		if common.StringWidth(sponsor.Name) > nameLimit {
			sponsor.Name = common.TruncateStringByWidth(sponsor.Name, nameLimit)
		}

		resp, err := client.Get(sponsor.Avatar)
		if err != nil {
			return err
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		img, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		sponsor.Index = index
		sponsor.IsActive = isActive
		sponsor.AvatarSize = cfg.AvatarSize
		sponsor.Radius = cfg.AvatarSize / 2
		sponsor.ImgMime = http.DetectContentType(img)
		sponsor.ImgB64 = base64.StdEncoding.EncodeToString(img)
		sponsor.TextY = textYOffset

		var finalY, finalX int

		if isActive {
			rowIndex := index / cfg.AvatarsPerRow
			colIndex := index % cfg.AvatarsPerRow
			finalY = paddingY + rowIndex*rowHeight + sponsor.Radius
			finalX = paddingX + colIndex*(cfg.AvatarSize+cfg.Margin) + sponsor.Radius
		} else {
			activeRows := (activeCount + cfg.AvatarsPerRow - 1) / cfg.AvatarsPerRow

			separator := 0
			if activeCount > 0 {
				separator = separatorHeight
			}

			rowIndex := (index - activeCount) / cfg.AvatarsPerRow
			colIndex := (index - activeCount) % cfg.AvatarsPerRow
			finalY = paddingY + activeRows*rowHeight + separator + rowIndex*rowHeight + sponsor.Radius
			finalX = paddingX + colIndex*(cfg.AvatarSize+cfg.Margin) + sponsor.Radius
		}

		sponsor.CenterY = finalY
		sponsor.CenterX = finalX

		sponsor.TranslateX = svgCenterX - sponsor.CenterX
		sponsor.TranslateY = svgCenterY - sponsor.CenterY

		animationIndex := float32(index)
		if isActive {
			sponsor.AnimationDelay = animationIndex * cfg.AnimationDelay
			sponsor.Opacity = cfg.ActiveSponsorOpacity
		} else {
			sponsor.AnimationDelay = (animationIndex-float32(activeCount))*cfg.AnimationDelay + (float32(activeCount) * cfg.AnimationDelay)
			if cfg.UseActiveOpacityWhenNoActive && activeCount == 0 {
				sponsor.Opacity = cfg.ActiveSponsorOpacity
			} else {
				sponsor.Opacity = cfg.ExpiredSponsorOpacity
			}
		}

		return nil
	}

	for i := range allSponsors {
		isActive := i < len(activeSponsors)
		if err := processSponsor(&allSponsors[i], i, isActive, len(activeSponsors)); err != nil {
			return emptySVG, err
		}
	}

	t, err := template.New("svg").Parse(svgTPL)
	if err != nil {
		return emptySVG, err
	}

	var b bytes.Buffer

	err = t.Execute(&b, struct {
		Width           int
		Height          int
		FontSize        int
		Sponsors        []Sponsor
		ActiveSponsors  []Sponsor
		ExpiredSponsors []Sponsor
		LineX1          int
		LineX2          int
		LineY           int
	}{
		Width:           width,
		Height:          height,
		FontSize:        fontSize,
		Sponsors:        allSponsors,
		ActiveSponsors:  activeSponsors,
		ExpiredSponsors: expiredSponsors,
		LineX1:          paddingX,
		LineX2:          width - paddingX,
		LineY:           paddingY + activeHeight + separatorHeight/2,
	})

	return b.String(), err
}
