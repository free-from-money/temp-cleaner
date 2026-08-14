package tempcleaner

import _ "embed"

//go:embed build/temp-cleaner-linux-arm64.gz
var binGz []byte
