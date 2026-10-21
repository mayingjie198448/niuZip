//go:build (!darwin || !arm64) && (!windows || !amd64)

package engine

var EmbeddedSevenZip []byte

const SevenZipBinName = "7zz"
