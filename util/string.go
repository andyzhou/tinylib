package util

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"unicode"
)

/*
 * string tools
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type String struct {
}

//remove html tag
func (f *String) RemoveHtmlTag(html string) string {
	re, _ := regexp.Compile("<.*?>")
	return re.ReplaceAllString(html, "")
}

//convert slice to string
func (f *String) Slice2Str(
	orgSlice []string,
	splits ...string) string {
	var (
		split, result string
		buffer *bytes.Buffer
	)
	if len(orgSlice) <= 0 {
		return result
	}
	if splits != nil && len(splits) > 0 {
		split = splits[0]
	}
	buffer = bytes.NewBuffer(nil)
	for idx, v := range orgSlice {
		buffer.WriteString(v)
		if idx > 0 && split != "" {
			buffer.WriteString(split)
		}
	}
	return buffer.String()
}

//sub string, support utf8 string
func (f *String) SubString(
	source string,
	start, length int) (string, error) {
	//check
	if source == "" {
		return source, errors.New("invalid parameter")
	}
	rs := []rune(source)
	rsLen := len(rs)
	if start < 0 {
		start = 0
	}
	if start >= rsLen {
		start = rsLen
	}
	end := start + length
	if end > rsLen {
		end = rsLen
	}
	return string(rs[start:end]), nil
}

//lower first character
func (f *String) LcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

//upper first character
func (f *String) UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

//verify string is english, numeric or combination
func (f *String) VerifyEnglishNumeric(input string) bool {
	if input == "" {
		return false
	}
	for _, v := range input {
		if !unicode.IsLetter(v) && !unicode.IsNumber(v) {
			return false
		}
	}
	return true
}

//remove html tags
func (f *String) TrimHtml(src string, needLowers ...bool) string {
	var (
		re *regexp.Regexp
		needLower bool
	)
	if needLowers != nil && len(needLowers) > 0 {
		needLower = needLowers[0]
	}

	if needLower {
		//convert to lower
		re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
		src = re.ReplaceAllStringFunc(src, strings.ToLower)
	}

	//remove style
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//remove script
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	return strings.TrimSpace(src)
}