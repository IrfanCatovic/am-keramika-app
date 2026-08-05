package utils

import (
	"regexp"
	"strings"
)

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

// GenerateSlug pretvara naziv u URL-friendly slug.
func GenerateSlug(name string) string {
	replacer := strings.NewReplacer(
		"š", "s", "đ", "dj", "č", "c", "ć", "c", "ž", "z",
		"Š", "s", "Đ", "dj", "Č", "c", "Ć", "c", "Ž", "z",
	)

	slug := replacer.Replace(strings.TrimSpace(name))
	slug = strings.ToLower(slug)
	slug = nonSlugChars.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return slug
}
