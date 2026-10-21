package cleaner_test

import (
	"testing"

	"github.com/mayingjie198448/niuZip/core/cleaner"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestDecodeZipFilename(t *testing.T) {
	// 1. Pure ASCII
	if got := cleaner.DecodeZipFilename("OpsHost/sub/main.go", true); got != "OpsHost/sub/main.go" {
		t.Errorf("expected OpsHost/sub/main.go, got %s", got)
	}

	// 2. Standard UTF-8 Chinese with nonUTF8: false
	utf8Chinese := "产品说明文档.docx"
	if got := cleaner.DecodeZipFilename(utf8Chinese, false); got != utf8Chinese {
		t.Errorf("expected %s, got %s", utf8Chinese, got)
	}

	// 3. Simplified Chinese GBK encoded bytes with nonUTF8: true
	gbkEncoder := simplifiedchinese.GB18030.NewEncoder()
	gbkBytes, err := gbkEncoder.Bytes([]byte("立达自动化设备_清单.xlsx"))
	if err != nil {
		t.Fatal(err)
	}
	rawGBK := string(gbkBytes)
	decoded := cleaner.DecodeZipFilename(rawGBK, true)
	if decoded != "立达自动化设备_清单.xlsx" {
		t.Errorf("expected '立达自动化设备_清单.xlsx', got '%s'", decoded)
	}

	// 4. GBK encoded with nonUTF8 incorrectly false (auto fallback due to invalid UTF-8)
	decodedAuto := cleaner.DecodeZipFilename(rawGBK, false)
	if decodedAuto != "立达自动化设备_清单.xlsx" {
		t.Errorf("expected auto recovery '立达自动化设备_清单.xlsx', got '%s'", decodedAuto)
	}

	// 5. Special Chinese word "图片" (whose GBK bytes happen to be valid UTF-8 Greek)
	tupianBytes, _ := gbkEncoder.Bytes([]byte("图片"))
	decodedTupian := cleaner.DecodeZipFilename(string(tupianBytes), true)
	if decodedTupian != "图片" {
		t.Errorf("expected '图片', got '%s'", decodedTupian)
	}

	// 6. Mixed Chinese and English deep path
	mixedBytes, _ := gbkEncoder.Bytes([]byte("项目源码/模块A_核心驱动/配置.json"))
	decodedMixed := cleaner.DecodeZipFilename(string(mixedBytes), true)
	if decodedMixed != "项目源码/模块A_核心驱动/配置.json" {
		t.Errorf("expected '项目源码/模块A_核心驱动/配置.json', got '%s'", decodedMixed)
	}

	// 7. Empty string
	if got := cleaner.DecodeZipFilename("", false); got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}
