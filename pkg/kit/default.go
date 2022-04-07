package kit

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// 生成消息摘要
func CreateDigest(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// 元素是否在其中
func IsContainInt(target int, raw []int) bool {
	sort.Ints(raw)
	index := sort.SearchInts(raw, target)
	if index < len(raw) && raw[index] == target {
		return true
	}
	return false
}

// 元素是否在其中
func IsContainString(target string, raw []string) bool {
	sort.Strings(raw)
	index := sort.SearchStrings(raw, target)
	if index < len(raw) && raw[index] == target {
		return true
	}
	return false
}

func Base64Decode(v string) (string, error) {
	temp, err := base64.StdEncoding.DecodeString(v)
	return string(temp), err
}

func Base64Encode(v string) string {
	temp := base64.StdEncoding.EncodeToString([]byte(v))
	return temp
}

// 字符串拼接
func JoinString(values ...string) string {
	var buffer strings.Builder
	for _, v := range values {
		buffer.WriteString(v)
	}
	return buffer.String()
}

// 链接拼接
func JoinLink(link, relatives string) (string, error) {
	u, err := url.Parse(relatives)
	if err != nil {
		return "", err
	}
	base, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(u).String(), nil
}
