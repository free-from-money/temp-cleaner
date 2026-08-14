package tempcleaner

import _ "embed"

//go:embed build/temp-cleaner-darwin-amd64.gz
var binGz []byte
