# 🌩 afdian-sponsor

> Generate afdian sponsor svg

[![GitHub License](https://img.shields.io/github/license/Sn0wo2/afdian-sponsor)](LICENSE)

[![Go CI](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml)
[![Release](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml)

## Usage

```yaml
name: afdian-sponsor manual run

on:
  workflow_dispatch:

jobs:
  build-and-run:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run afdian-sponsor action
        uses: Sn0wo2/afdian-sponsor@v1.0.0 # replace with the latest version
        env:
          AFDIAN_USER_ID: ${{ secrets.AFDIAN_USER_ID }}
          AFDIAN_API_TOKEN: ${{ secrets.AFDIAN_API_TOKEN }}

      - name: Upload generated SVG
        uses: actions/upload-artifact@v4
        with:
          name: afdian-sponsor-svg
          path: afdian-sponsor.svg
```

## 📄 **License**

Licensed under [GPL 3.0](LICENSE).