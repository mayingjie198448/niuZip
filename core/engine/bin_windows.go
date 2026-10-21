//go:build windows && amd64

package engine

import _ "embed"

//go:embed bin/windows_amd64/7zz.exe
var EmbeddedSevenZip []byte

const SevenZipBinName = "7zz.exe"
