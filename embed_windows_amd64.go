package tempcleaner

import _ "embed"

//go:embed build/temp-cleaner-windows-amd64.exe.gz
var binGz []byte
