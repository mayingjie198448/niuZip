//go:build darwin && arm64

package engine

import _ "embed"

//go:embed bin/darwin_arm64/7zz
var EmbeddedSevenZip []byte

const SevenZipBinName = "7zz"
