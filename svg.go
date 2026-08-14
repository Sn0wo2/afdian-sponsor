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
	"net/url"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/mattn/go-runewidth"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	minifysvg "github.com/tdewolff/minify/v2/svg"
	"golang.org/x/image/draw"
)

//go:embed assets/templates/sponsors.svg
var SVGTemplateSource string

const (
	EmptySVG          = `<svg xmlns="http://www.w3.org/2000/svg" width="1135" height="100"/>`
	GoogleFontsCSSURL = "https://fonts.googleapis.com/css2"
	UA                = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/131.0 Safari/537.36"

	MaxGoogleCSSBytes = 256 << 10
	MaxFontBytes      = 4 << 20
	MaxAvatarBytes    = 20 << 20
	MaxAvatarPixels   = 16_000_000
	JpegQuality       = 85
	MinLossySavings   = 10
)

var GoogleFontURLPattern = regexp.MustCompile(`url\((?:'|")?(https://fonts\.gstatic\.com/[^)'"]+)`)

var SVGMinifier = func() *minify.M {
	minifier := minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("image/svg+xml", minifysvg.Minify)

	return minifier
}()

var SponsorsSVGTemplate = template.Must(template.New("svg").Funcs(template.FuncMap{
	"xml": func(value string) string {
		var escaped bytes.Buffer

		_ = xml.EscapeText(&escaped, []byte(value))

		return escaped.String()
	},
}).Parse(SVGTemplateSource))

type SVGRenderer struct {
	client *http.Client
	config Config
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

type SVGDocument struct {
	Width         int
	Height        int
	FontSize      int
	FontBase64    string
	FontFallback  string
	AvatarRadius  int
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
		FontFallback:  strings.SplitN(config.FontFamily, ":", 2)[0],
		AvatarRadius:  config.AvatarSize / 2,
		Sponsors:      make([]SVGSponsor, 0, len(groups.active)+len(groups.expired)),
		ShowSeparator: separatorHeight > 0,
		LineX1:        paddingX,
		LineX2:        width - paddingX,
		LineY:         paddingY + activeHeight + separatorHeight/2,
	}

	appendSponsors := func(sponsors []Sponsor, active bool, yOffset int) {
		for groupIndex, sponsor := range sponsors {
			avatar, err := renderer.FetchAvatar(sponsor.Avatar)
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

	seenRunes := make(map[rune]struct{})
	var fontText strings.Builder
	for _, sponsor := range document.Sponsors {
		for _, character := range sponsor.Name {
			if _, seen := seenRunes[character]; seen {
				continue
			}

			seenRunes[character] = struct{}{}
			fontText.WriteRune(character)
		}
	}

	if fontText.Len() > 0 {
		font, err := renderer.FetchGoogleFontSubset(fontText.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: unable to embed Google Fonts subset: %v\n", err)
		} else {
			document.FontBase64 = base64.StdEncoding.EncodeToString(font)
		}
	}

	var output bytes.Buffer
	if err := SponsorsSVGTemplate.Execute(&output, document); err != nil {
		return EmptySVG, fmt.Errorf("render SVG template: %w", err)
	}

	minified, err := SVGMinifier.String("image/svg+xml", output.String())
	if err != nil {
		return EmptySVG, fmt.Errorf("minify SVG: %w", err)
	}

	return minified, nil
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
	case renderer.config.FontFamily == "":
		return errors.New("font family must not be empty")
	case renderer.config.PaddingXScale < 0 || renderer.config.PaddingYScale < 0:
		return errors.New("padding scales cannot be negative")
	default:
		return nil
	}
}

func (renderer SVGRenderer) FetchGoogleFontSubset(text string) ([]byte, error) {
	query := url.Values{}
	query.Set("display", "swap")
	query.Set("family", renderer.config.FontFamily)
	query.Set("text", text)

	stylesheet, err := renderer.FetchResource(GoogleFontsCSSURL+"?"+query.Encode(), MaxGoogleCSSBytes)
	if err != nil {
		return nil, fmt.Errorf("fetch stylesheet: %w", err)
	}

	match := GoogleFontURLPattern.FindSubmatch(stylesheet)
	if len(match) != 2 {
		return nil, errors.New("stylesheet does not contain a WOFF2 URL")
	}

	fontURL, err := url.Parse(string(match[1]))
	if err != nil || fontURL.Scheme != "https" || fontURL.Hostname() != "fonts.gstatic.com" {
		return nil, errors.New("stylesheet contains an invalid font URL")
	}

	font, err := renderer.FetchResource(fontURL.String(), MaxFontBytes)
	if err != nil {
		return nil, fmt.Errorf("fetch font: %w", err)
	}
	if len(font) < 4 || string(font[:4]) != "wOF2" {
		return nil, errors.New("font response is not WOFF2")
	}

	return font, nil
}

func (renderer SVGRenderer) FetchAvatar(resourceURL string) ([]byte, error) {
	avatar, err := renderer.FetchResource(resourceURL, MaxAvatarBytes)
	if err != nil {
		return nil, err
	}

	return renderer.OptimizeAvatar(avatar), nil
}

func (renderer SVGRenderer) FetchResource(resourceURL string, maxBytes int64) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, resourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("User-Agent", UA)

	response, err := renderer.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4<<10))

		return nil, fmt.Errorf("unexpected HTTP status %s", response.Status)
	}
	if response.ContentLength > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("empty response")
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}

	return data, nil
}

func (renderer SVGRenderer) OptimizeAvatar(source []byte) []byte {
	mimeType := http.DetectContentType(source)
	if mimeType != "image/jpeg" && mimeType != "image/png" || mimeType == "image/png" && isAnimatedPNG(source) {
		return source
	}

	imageConfig, _, err := image.DecodeConfig(bytes.NewReader(source))
	if err != nil || imageConfig.Width <= 0 || imageConfig.Height <= 0 || imageConfig.Height > MaxAvatarPixels/imageConfig.Width {
		return source
	}

	decoded, _, err := image.Decode(bytes.NewReader(source))
	if err != nil {
		return source
	}

	optimized := decoded
	if imageConfig.Width > renderer.config.AvatarSize || imageConfig.Height > renderer.config.AvatarSize {
		width, height := imageConfig.Width, imageConfig.Height
		if width >= height {
			width = renderer.config.AvatarSize
			height = max(1, imageConfig.Height*renderer.config.AvatarSize/imageConfig.Width)
		} else {
			height = renderer.config.AvatarSize
			width = max(1, imageConfig.Width*renderer.config.AvatarSize/imageConfig.Height)
		}

		resized := image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(resized, resized.Bounds(), decoded, decoded.Bounds(), draw.Src, nil)
		optimized = resized
	}

	best := source
	var pngOutput bytes.Buffer
	pngEncoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := pngEncoder.Encode(&pngOutput, optimized); err == nil && pngOutput.Len() < len(best) {
		best = pngOutput.Bytes()
	}

	if mimeType == "image/png" && !imageIsOpaque(optimized) {
		return best
	}

	var jpegOutput bytes.Buffer
	if err := jpeg.Encode(&jpegOutput, optimized, &jpeg.Options{Quality: JpegQuality}); err == nil && jpegOutput.Len() <= len(best)*(100-MinLossySavings)/100 {
		best = jpegOutput.Bytes()
	}

	return best
}

func imageIsOpaque(source image.Image) bool {
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

func isAnimatedPNG(source []byte) bool {
	if len(source) < 8 || !bytes.Equal(source[:8], []byte("\x89PNG\r\n\x1a\n")) {
		return false
	}

	for offset := 8; offset+12 <= len(source); {
		length := uint64(binary.BigEndian.Uint32(source[offset : offset+4]))
		if length > uint64(len(source)-offset-12) {
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
