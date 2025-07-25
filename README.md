# 🌩 afdian-sponsor

> Generate afdian sponsors svg on github action

[![Go Report Card](https://goreportcard.com/badge/github.com/Sn0wo2/afdian-sponsor)](https://goreportcard.com/report/github.com/Sn0wo2/afdian-sponsor)
[![GitHub release](https://img.shields.io/github/v/release/Sn0wo2/afdian-sponsor?color=blue)](https://github.com/Sn0wo2/afdian-sponsor/releases)
[![GitHub License](https://img.shields.io/github/license/Sn0wo2/afdian-sponsor)](LICENSE)

[![Go CI](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml)
[![Release](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml)

## 👀 Example

See **https://github.com/Sn0wo2/Sn0wo2/blob/main/.github/workflows/sponsor.yml**

[![](https://github.com/Sn0wo2/Sn0wo2/raw/refs/heads/out/sponsor/afdian-sponsor.svg)](https://afdian.com/a/Me0wo)

## 🚀 Usage

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

      - name: Run afdian-sponsor action
        uses: Sn0wo2/afdian-sponsor@v1
        with:
          # Optional(default values)
          version: latest
          cache: true
        env:
          # Required
          # Get User ID and API Token from https://afdian.com/dashboard/dev
          # And add them to your github secrets(https://github.com/your-user-name/your-repo/settings/secrets/actions/new)
          AFDIAN_USER_ID: ${{ secrets.AFDIAN_USER_ID }}
          AFDIAN_API_TOKEN: ${{ secrets.AFDIAN_API_TOKEN }}

          # Optional(default values)
          AFDIAN_OUTPUT: ./
          AFDIAN_TOTAL_SPONSORS: 100
          AFDIAN_AVATAR_SIZE: 100
          AFDIAN_MARGIN: 15
          AFDIAN_AVATARS_PER_ROW: 15

      - name: Upload generated SVG
        uses: actions/upload-artifact@v4
        with:
          name: afdian-sponsor-svg
          path: afdian-sponsor.svg
```

## 🔗 Links

- [go-afdian-api](https://github.com/Sn0wo2/go-afdian-api)