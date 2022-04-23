Mobile TinyMan SDK for golang.

## Architecture
1. There's a community driven Pool.json file that needs to be updated by the community. This file contains all verified addresses of TinyMan AMM pools. The Format of the json
```json
 "pools": [
        {
            "key":"",
            "address": ""
        }
    ]
```
the key as the format of {min(assetID)}-{max(assetID)}

2. This sdk written in golang. for each update on the pool.json, You are to use
[go-bind](https://pkg.go.dev/golang.org/x/mobile/cmd/gobind) to generate Android and IOS packages.

