# afdian-sponsor

> Generate [afdian](https://afdian.com) sponsors svg on github action

[![Go Report Card](https://goreportcard.com/badge/github.com/Sn0wo2/afdian-sponsor)](https://goreportcard.com/report/github.com/Sn0wo2/afdian-sponsor)
[![GitHub release](https://img.shields.io/github/v/release/Sn0wo2/afdian-sponsor?color=blue)](https://github.com/Sn0wo2/afdian-sponsor/releases)
[![GitHub License](https://img.shields.io/github/license/Sn0wo2/afdian-sponsor)](LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2Fafdian-sponsor.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2Fafdian-sponsor?ref=badge_shield)

[![Go CI](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml)
[![Release](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml)

---

## Demo

See **https://github.com/Sn0wo2/Sn0wo2/blob/main/.github/workflows/sponsor.yml**

[![](https://github.com/Sn0wo2/Sn0wo2/raw/refs/heads/out/sponsor/afdian-sponsor.svg)](https://afdian.com/a/Me0wo)

## Example

```yaml
name: Sponsor

on:
  workflow_dispatch:

jobs:
  build-and-run:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run ifdian-sponsor action
        uses: Sn0wo2/afdian-sponsor@v1
        with:
          # Optional(default values)
          version: latest
          cache: true
        env:
          # Required
          # Get User ID and API Token from https://ifdian.net/dashboard/dev
          # And add them to your github secrets(https://github.com/$your-user-name/$your-repo/settings/secrets/actions/new)
          AFDIAN_USER_ID: ${{ secrets.AFDIAN_USER_ID }}
          AFDIAN_API_TOKEN: ${{ secrets.AFDIAN_API_TOKEN }}

          # Optional(default values)
          AFDIAN_OUTPUT: ./afdian-sponsor.svg
          AFDIAN_TOTAL_SPONSORS: 100
          AFDIAN_AVATAR_SIZE: 100
          AFDIAN_MARGIN: 15
          AFDIAN_AVATARS_PER_ROW: 15
          # Google Fonts CSS2 family value, e.g. "Noto Sans SC:wght@100..900" (default) or "ZCOOL QingKe HuangYou"
          # When set, the action fetches the matching font itself and hands it to the binary via AFDIAN_FONT_FILE;
          AFDIAN_FONT_FAMILY: Noto Sans SC:wght@100..900
          AFDIAN_FONT_FILE: ./font.ttf
          AFDIAN_FONTSIZE_SCALE: 8
          AFDIAN_PADDINGX_SCALE: 2
          AFDIAN_PADDINGY_SCALE: 4
          AFDIAN_SORT: time

      - name: Upload generated SVG
        uses: actions/upload-artifact@v4
        with:
          name: afdian-sponsor-svg
          path: afdian-sponsor.svg
```

## Fonts

The binary never fetches fonts at runtime. The default font (`Noto Sans SC:wght@100..900`) is fetched at build time
(`make font` / `go generate ./internal/font`, implemented by `scripts/font`) and embedded.
`scripts/font/charset.txt` defines the character subset (the 500 most common simplified Chinese characters plus ASCII,
digits and punctuation, from [linkary/top-used-chars](https://github.com/linkary/top-used-chars), MIT), constrained by
Google Fonts' ~700 character subset limit.

When `AFDIAN_FONT_FAMILY` is set, the action fetches the matching subset itself and passes it to the binary via
`AFDIAN_FONT_FILE`; when unset, the binary uses its embedded font. Characters outside the subset fall back to the system
font.

--- 

> [go-afdian-api](https://github.com/Sn0wo2/go-afdian-api)

---

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2Fafdian-sponsor.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2Fafdian-sponsor?ref=badge_large)
