# afdian-sponsor

> Generate [ifdian(afdian)](https://ifdian.net) sponsors svg on github action

[![Go Report Card](https://goreportcard.com/badge/github.com/Sn0wo2/afdian-sponsor)](https://goreportcard.com/report/github.com/Sn0wo2/afdian-sponsor)
[![GitHub release](https://img.shields.io/github/v/release/Sn0wo2/afdian-sponsor?color=blue)](https://github.com/Sn0wo2/afdian-sponsor/releases)
[![GitHub License](https://img.shields.io/github/license/Sn0wo2/afdian-sponsor)](LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2Fafdian-sponsor.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2Fafdian-sponsor?ref=badge_shield)

[![Go CI](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/go.yml)
[![Release](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml/badge.svg)](https://github.com/Sn0wo2/afdian-sponsor/actions/workflows/release.yml)

---

不知何时**爱发电**已经把`afdian.com`的域名重定向到`ifdian.net`, 旧域名仍然可以提供服务,  
但是因为此项目是 _github action_ 项目名将仍然保持`afdian-sponsor`  
我们将项目中使用到的`afdian.com`的域名逐步替换到`ifdian.net`, 在描述中我们也将以**ifdian**指代之前的**afdian**

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
          AFDIAN_OUTPUT: ./
          AFDIAN_TOTAL_SPONSORS: 100
          AFDIAN_AVATAR_SIZE: 100
          AFDIAN_MARGIN: 15
          AFDIAN_AVATARS_PER_ROW: 15
          AFDIAN_FONTSIZE_SCALE: 8
          AFDIAN_PADDINGX_SCALE: 2
          AFDIAN_PADDINGY_SCALE: 4
          AFDIAN_SORT: time

      - name: Upload generated SVG
        uses: actions/upload-artifact@v4
        with:
          name: ifdian-sponsor-svg
          path: ifdian-sponsor.svg
```

--- 

> [go-afdian-api](https://github.com/Sn0wo2/go-afdian-api)

---

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FSn0wo2%2Fafdian-sponsor.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FSn0wo2%2Fafdian-sponsor?ref=badge_large)