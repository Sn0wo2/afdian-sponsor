# 🌩 afdian-sponsor

> Generate afdian sponsors svg on github action

[![Go Report Card](https://goreportcard.com/badge/github.com/Sn0wo2/afdian-sponsor)](https://goreportcard.com/report/github.com/Sn0wo2/afdian-sponsor)
[![GitHub release](https://img.shields.io/github/v/release/Sn0wo2/afdian-sponsor?color=blue)](https://github.com/Sn0wo2/afdian-sponsor/releases)
[![GitHub License](https://img.shields.io/github/license/Sn0wo2/afdian-sponsor)](LICENSE)

[![Go CI](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml)
[![Release](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml)

## 👀 Example

See **https://github.com/Sn0wo2/Sn0wo2**

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
        env:
          # Required
          # Get User ID and API Token from https://afdian.com/dashboard/dev
          AFDIAN_USER_ID: ${{ secrets.AFDIAN_USER_ID }}
          AFDIAN_API_TOKEN: ${{ secrets.AFDIAN_API_TOKEN }}

          # Optional(default values)
          AFDIAN_OUTPUT: /github/workspace/afdian-sponsor.svg
          AFDIAN_PAGE: 1
          AFDIAN_PER_PAGE: 100
          AFDIAN_AVATAR_SIZE: 100
          AFDIAN_MARGIN: 15
          AFDIAN_AVATARS_PER_ROW: 10

      - name: Upload generated SVG
        uses: actions/upload-artifact@v4
        with:
          name: afdian-sponsor-svg
          path: afdian-sponsor.svg
```

## 📄 **License**

Licensed under [GPL 3.0](LICENSE).