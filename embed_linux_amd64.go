package tempcleaner

import _ "embed"

//go:embed build/temp-cleaner-linux-amd64.gz
var binGz []byte
