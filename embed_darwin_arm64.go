package tempcleaner

import _ "embed"

//go:embed build/temp-cleaner-darwin-arm64.gz
var binGz []byte
