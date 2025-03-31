package testdata

import (
	_ "embed"
)

//go:embed manifest.json
var ManifestJSON []byte

//go:embed manifest-app-name.json
var ManifestJSONAppName []byte

//go:embed manifest.js
var ManifestJS []byte

//go:embed manifest-app-name.js
var ManifestJSAppName []byte

//go:embed manifest.ts
var ManifestTS []byte

//go:embed manifest-app-name.ts
var ManifestTSAppName []byte

//go:embed manifest-sdk.ts
var ManifestSDKTS []byte

//go:embed manifest-sdk-app-name.ts
var ManifestSDKTSAppName []byte
