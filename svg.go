package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/mattn/go-runewidth"
)

//go:embed assets/fonts/SourceHanSansCN-VF.ttf.woff2
var SourceHanSansFont []byte

//go:embed assets/templates/sponsors.svg
var SVGTemplateSource string

var SourceHanSansFontBase64 = base64.StdEncoding.EncodeToString(SourceHanSansFont)

const EmptySVG = `<svg width="1135" height="100" xmlns="http://www.w3.org/2000/svg" style="background-color:transparent;"></svg>`

type SVGRenderer struct {
	client *http.Client
	config Config
}

type SVGSponsor struct {
	Name           string
	OriginalName   string
	Index          int
	CenterX        int
	CenterY        int
	TextY          int
	Radius         int
	AvatarSize     int
	ImageMIME      string
	ImageBase64    string
	AnimationDelay float32
	Opacity        float32
	Active         bool
	TranslateX     int
	TranslateY     int
}

type SVGDocument struct {
	Width         int
	Height        int
	FontSize      int
	FontBase64    string
	Sponsors      []SVGSponsor
	ShowSeparator bool
	LineX1        int
	LineX2        int
	LineY         int
}

func NewSVGRenderer(client *http.Client, config Config) SVGRenderer {
	return SVGRenderer{client: client, config: config}
}

func (renderer SVGRenderer) Render(groups SponsorGroups) (string, error) {
	if len(groups.active) == 0 && len(groups.expired) == 0 {
		return EmptySVG, nil
	}

	if err := renderer.Validate(); err != nil {
		return EmptySVG, err
	}

	config := renderer.config
	fontSize := max(1, config.AvatarSize/config.FontSizeScale)

	nameLimit := max(5, config.AvatarSize*2/fontSize)

	paddingX := 0
	if config.PaddingXScale > 0 {
		paddingX = config.AvatarSize / config.PaddingXScale
	}

	paddingY := 0
	if config.PaddingYScale > 0 {
		paddingY = config.AvatarSize / config.PaddingYScale
	}

	rowHeight := config.AvatarSize + config.Margin + fontSize + 10
	textY := config.AvatarSize/2 + fontSize + 10

	activeRows := (len(groups.active) + config.AvatarsPerRow - 1) / config.AvatarsPerRow
	activeHeight := activeRows * rowHeight

	separatorHeight := 0
	if len(groups.active) > 0 && len(groups.expired) > 0 {
		separatorHeight = 40
	}

	expiredRows := (len(groups.expired) + config.AvatarsPerRow - 1) / config.AvatarsPerRow
	expiredHeight := expiredRows * rowHeight
	height := paddingY + activeHeight + separatorHeight + expiredHeight
	width := config.AvatarsPerRow*(config.AvatarSize+config.Margin) - config.Margin + paddingX*2
	centerX := width / 2
	centerY := height / 2

	document := SVGDocument{
		Width:         width,
		Height:        height,
		FontSize:      fontSize,
		FontBase64:    SourceHanSansFontBase64,
		Sponsors:      make([]SVGSponsor, 0, len(groups.active)+len(groups.expired)),
		ShowSeparator: separatorHeight > 0,
		LineX1:        paddingX,
		LineX2:        width - paddingX,
		LineY:         paddingY + activeHeight + separatorHeight/2,
	}

	appendSponsors := func(sponsors []Sponsor, active bool, yOffset int) {
		for groupIndex, sponsor := range sponsors {
			image, err := renderer.FetchAvatar(sponsor.Avatar)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: skipping avatar for sponsor %q: %v\n", sponsor.Name, err)
			}

			name := sponsor.Name
			if runewidth.StringWidth(name) > nameLimit {
				name = runewidth.Truncate(name, nameLimit, "") + "..."
			}

			row := groupIndex / config.AvatarsPerRow
			column := groupIndex % config.AvatarsPerRow
			radius := config.AvatarSize / 2
			x := paddingX + column*(config.AvatarSize+config.Margin) + radius
			y := yOffset + row*rowHeight + radius

			opacity := config.ExpiredSponsorOpacity
			if active || len(groups.active) == 0 && config.UseActiveOpacityWhenNoActive {
				opacity = config.ActiveSponsorOpacity
			}

			index := len(document.Sponsors)

			renderSponsor := SVGSponsor{
				Name:           name,
				OriginalName:   sponsor.Name,
				Index:          index,
				CenterX:        x,
				CenterY:        y,
				TextY:          textY,
				Radius:         radius,
				AvatarSize:     config.AvatarSize,
				AnimationDelay: float32(index) * config.AnimationDelay,
				Opacity:        opacity,
				Active:         active,
				TranslateX:     centerX - x,
				TranslateY:     centerY - y,
			}
			if len(image) > 0 {
				renderSponsor.ImageMIME = http.DetectContentType(image)
				renderSponsor.ImageBase64 = base64.StdEncoding.EncodeToString(image)
			}

			document.Sponsors = append(document.Sponsors, renderSponsor)
		}
	}

	appendSponsors(groups.active, true, paddingY)
	appendSponsors(groups.expired, false, paddingY+activeHeight+separatorHeight)

	var output bytes.Buffer
	if err := template.Must(template.New("svg").Funcs(template.FuncMap{
		"xml": func(value string) string {
			var escaped bytes.Buffer

			_ = xml.EscapeText(&escaped, []byte(value))

			return escaped.String()
		},
	}).Parse(SVGTemplateSource)).Execute(&output, document); err != nil {
		return EmptySVG, fmt.Errorf("render SVG template: %w", err)
	}

	return output.String(), nil
}

func (renderer SVGRenderer) Validate() error {
	switch {
	case renderer.client == nil:
		return errors.New("HTTP client is required")
	case renderer.config.AvatarSize <= 0:
		return errors.New("avatar size must be positive")
	case renderer.config.Margin < 0:
		return errors.New("margin cannot be negative")
	case renderer.config.AvatarsPerRow <= 0:
		return errors.New("avatars per row must be positive")
	case renderer.config.FontSizeScale <= 0:
		return errors.New("font size scale must be positive")
	case renderer.config.PaddingXScale < 0 || renderer.config.PaddingYScale < 0:
		return errors.New("padding scales cannot be negative")
	default:
		return nil
	}
}

func (renderer SVGRenderer) FetchAvatar(url string) ([]byte, error) {
	response, err := renderer.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, response.Body)

		return nil, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}

	image, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if len(image) == 0 {
		return nil, errors.New("empty response")
	}

	return image, nil
}
