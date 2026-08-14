package tempcleaner

import _ "embed"

//go:embed build/temp-cleaner-windows-arm64.exe.gz
var binGz []byte
