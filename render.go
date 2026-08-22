package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/Sn0wo2/afdian-sponsor/internal/font"
	"github.com/gabriel-vasile/mimetype"
	"github.com/mattn/go-runewidth"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	minifysvg "github.com/tdewolff/minify/v2/svg"
	"golang.org/x/image/draw"
)

//go:embed assets/templates/sponsors.svg
var SVGTemplateSource string

//go:embed assets/templates/empty.svg
var EmptySVG string

const (
	MaxAvatarBytes = 10 * 1024 * 1024
)

var SponsorsSVGTemplate = template.Must(template.New("svg").Funcs(template.FuncMap{
	"xml": func(value string) string {
		var escaped bytes.Buffer

		_ = xml.EscapeText(&escaped, []byte(value))

		return escaped.String()
	},
}).Parse(SVGTemplateSource))

type SVGDocument struct {
	Width         int
	Height        int
	FontSize      int
	FontBase64    string
	FontMIME      string
	FontFallback  string
	AvatarRadius  int
	Sponsors      []SVGSponsor
	ShowSeparator bool
	LineX1        int
	LineX2        int
	LineY         int
}

type SVGSponsor struct {
	Name           string
	OriginalName   string
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

func Render(config Config, groups SponsorGroups) (string, error) {
	if len(groups.active) == 0 && len(groups.expired) == 0 {
		return EmptySVG, nil
	}

	switch {
	case config.AvatarSize <= 0:
		return EmptySVG, errors.New("avatar size must be positive")
	case config.Margin < 0:
		return EmptySVG, errors.New("margin cannot be negative")
	case config.AvatarsPerRow <= 0:
		return EmptySVG, errors.New("avatars per row must be positive")
	case config.FontSizeScale <= 0:
		return EmptySVG, errors.New("font size scale must be positive")
	case config.PaddingXScale < 0 || config.PaddingYScale < 0:
		return EmptySVG, errors.New("padding scales cannot be negative")
	}

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
		FontFallback:  "Noto Sans SC",
		AvatarRadius:  config.AvatarSize / 2,
		Sponsors:      make([]SVGSponsor, 0, len(groups.active)+len(groups.expired)),
		ShowSeparator: separatorHeight > 0,
		LineX1:        paddingX,
		LineX2:        width - paddingX,
		LineY:         paddingY + activeHeight + separatorHeight/2,
	}

	isAnimatedPNG := func(source []byte) bool {
		if len(source) < 8 || !bytes.Equal(source[:8], []byte("\x89PNG\r\n\x1a\n")) {
			return false
		}

		for offset := 8; offset+12 <= len(source); {
			length := binary.BigEndian.Uint32(source[offset : offset+4])
			if int64(length) > int64(len(source)-offset-12) {
				return false
			}

			chunkType := string(source[offset+4 : offset+8])
			if chunkType == "acTL" {
				return true
			}

			if chunkType == "IEND" {
				return false
			}

			offset += int(length) + 12
		}

		return false
	}

	imageIsOpaque := func(source image.Image) bool {
		if opaque, ok := source.(interface{ Opaque() bool }); ok {
			return opaque.Opaque()
		}

		bounds := source.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				_, _, _, alpha := source.At(x, y).RGBA()
				if alpha != 0xffff {
					return false
				}
			}
		}

		return true
	}

	appendSponsors := func(sponsors []Sponsor, active bool, yOffset int) {
		for groupIndex, sponsor := range sponsors {
			var avatar []byte

			resp, err := http.Get(sponsor.Avatar)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: skipping avatar for sponsor %q: create request: %v\n", sponsor.Name, err)
			} else {
				if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
					// 读完数据,让连接有机会被复用,但是最多4kb
					_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4*1024))

					fmt.Fprintf(os.Stderr, "warning: skipping avatar for sponsor %q: unexpected HTTP status %s\n", sponsor.Name, resp.Status)
				} else if resp.ContentLength > MaxAvatarBytes {
					fmt.Fprintf(os.Stderr, "warning: skipping avatar for sponsor %q: response exceeds %d bytes\n", sponsor.Name, MaxAvatarBytes)
				} else if data, readErr := io.ReadAll(io.LimitReader(resp.Body, MaxAvatarBytes+1)); readErr != nil {
					fmt.Fprintf(os.Stderr, "warning: skipping avatar for sponsor %q: read response: %v\n", sponsor.Name, readErr)
				} else if len(data) > MaxAvatarBytes {
					// ContentLength 为 -1 时上面的检查不生效,读满说明超限,跳过而不是嵌入截断的图片
					fmt.Fprintf(os.Stderr, "warning: skipping avatar for sponsor %q: response exceeds %d bytes\n", sponsor.Name, MaxAvatarBytes)
				} else {
					avatar = data

					mimeType := http.DetectContentType(data)
					if mimeType == "image/jpeg" || mimeType == "image/png" {
						if !(mimeType == "image/png" && isAnimatedPNG(data)) {
							if imageConfig, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil && imageConfig.Width > 0 && imageConfig.Height > 0 && imageConfig.Height <= 16_000_000/imageConfig.Width {
								if decoded, _, err := image.Decode(bytes.NewReader(data)); err == nil {
									if imageConfig.Width > config.AvatarSize || imageConfig.Height > config.AvatarSize {
										var width, height int
										if imageConfig.Width >= imageConfig.Height {
											width = max(config.AvatarSize, imageConfig.Width*config.AvatarSize/imageConfig.Height)
											height = config.AvatarSize
										} else {
											width = config.AvatarSize
											height = max(config.AvatarSize, imageConfig.Height*config.AvatarSize/imageConfig.Width)
										}

										resized := image.NewNRGBA(image.Rect(0, 0, width, height))
										draw.CatmullRom.Scale(resized, resized.Bounds(), decoded, decoded.Bounds(), draw.Src, nil)

										cropX := (width - config.AvatarSize) / 2
										cropY := (height - config.AvatarSize) / 2
										decoded = resized.SubImage(image.Rect(cropX, cropY, cropX+config.AvatarSize, cropY+config.AvatarSize))
									}

									var pngOutput bytes.Buffer

									pngEncoder := png.Encoder{CompressionLevel: png.BestCompression}
									if err := pngEncoder.Encode(&pngOutput, decoded); err == nil && pngOutput.Len() < len(avatar) {
										avatar = pngOutput.Bytes()
									}

									if !(mimeType == "image/png" && !imageIsOpaque(decoded)) {
										var jpegOutput bytes.Buffer
										if err := jpeg.Encode(&jpegOutput, decoded, &jpeg.Options{Quality: 85}); err == nil && jpegOutput.Len() <= len(avatar)*(100-10)/100 {
											avatar = jpegOutput.Bytes()
										}
									}
								}
							}
						}
					}
				}

				_ = resp.Body.Close()
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
			if len(avatar) > 0 {
				renderSponsor.ImageMIME = http.DetectContentType(avatar)
				renderSponsor.ImageBase64 = base64.StdEncoding.EncodeToString(avatar)
			}

			document.Sponsors = append(document.Sponsors, renderSponsor)
		}
	}

	appendSponsors(groups.active, true, paddingY)
	appendSponsors(groups.expired, false, paddingY+activeHeight+separatorHeight)

	fontData, err := font.Font, error(nil)
	if config.FontFile != "" {
		fontData, err = os.ReadFile(config.FontFile)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: unable to load font: %v\n", err)
	} else {
		document.FontBase64 = base64.StdEncoding.EncodeToString(fontData)
		document.FontMIME = mimetype.Detect(fontData).String()
	}

	var output bytes.Buffer
	if err := SponsorsSVGTemplate.Execute(&output, document); err != nil {
		return EmptySVG, fmt.Errorf("render SVG template: %w", err)
	}

	minifier := minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("image/svg+xml", minifysvg.Minify)

	minified, err := minifier.String("image/svg+xml", output.String())
	if err != nil {
		return EmptySVG, fmt.Errorf("minify SVG: %w", err)
	}

	return minified, nil
}
