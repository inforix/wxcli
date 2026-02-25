package markup

import "strings"

// UnescapeStyleEntities removes HTML entity escaping within style attributes
// to improve compatibility with downstream HTML sanitizers.
func UnescapeStyleEntities(html string) string {
	const key = `style="`
	var b strings.Builder
	idx := 0
	for {
		pos := strings.Index(html[idx:], key)
		if pos < 0 {
			b.WriteString(html[idx:])
			break
		}
		pos += idx
		b.WriteString(html[idx : pos+len(key)])
		start := pos + len(key)
		end := start
		for end < len(html) && html[end] != '"' {
			end++
		}
		if end > start {
			value := html[start:end]
			value = strings.ReplaceAll(value, "&#39;", "'")
			value = strings.ReplaceAll(value, "&apos;", "'")
			b.WriteString(value)
		}
		if end < len(html) {
			b.WriteByte('"')
			idx = end + 1
		} else {
			idx = end
		}
	}
	return b.String()
}
