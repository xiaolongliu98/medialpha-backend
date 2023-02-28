package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestFormatWindowsPathAbs(t *testing.T) {
	testExamples := map[string]string{
		"":             "",
		"D:/":          "D:",
		"D:\\Videos\\": "D:/Videos",
		"/":            "",
		"\\":           "",
	}

	for query, ans := range testExamples {

		res := FormatWindowsPathAbs(query)

		if res != ans {
			t.Errorf("[error] '%v' -> '%v', but answer is: %v", query, res, ans)
		}
	}
}

func TestFormatUnixPathAbs(t *testing.T) {
	testExamples := map[string]string{
		"":       "/",
		"a":      "/a",
		"/a/b/c": "/a/b/c",
		"/a/":    "/a",
		"/a/b":   "/a/b",
		"a/b/c":  "/a/b/c",
		"a/":     "/a",
	}
	// "a", ""
	// "/a/b/c/", "/a/", "/a/b", "a/b/c", "a/"
	for query, ans := range testExamples {

		res := FormatUnixPathAbs(query)

		if res != ans {
			t.Errorf("[error] '%v' -> '%v', but answer is: %v", query, res, ans)
		}
	}
}

func TestCompress(t *testing.T) {
	target := "D:/Videos/咒怨sknjbnij@$!!()($*%(((()%()^sdfl.qwoob"
	res := base64.StdEncoding.EncodeToString([]byte(target))
	strings.ReplaceAll(res, "/", "-")
	fmt.Println(res)

	res = strings.ReplaceAll(res, "-", "/")
	resByts, _ := base64.StdEncoding.DecodeString(res)
	res = string(resByts)

	fmt.Println(res)
}
