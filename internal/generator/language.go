package generator

import "fmt"

// Language represents a programming language target
type Language string

const (
	LanguageGo     Language = "go"
	LanguageDotNet Language = "dotnet"
	LanguageNodeJS Language = "nodejs"
)

// Valid languages
var validLanguages = []Language{
	LanguageGo,
	LanguageDotNet,
	LanguageNodeJS,
}

// IsValid checks if a language is valid
func (l Language) IsValid() bool {
	for _, valid := range validLanguages {
		if l == valid {
			return true
		}
	}
	return false
}

// String returns the string representation
func (l Language) String() string {
	return string(l)
}

// ValidLanguages returns all valid languages
func ValidLanguages() []string {
	result := make([]string, len(validLanguages))
	for i, lang := range validLanguages {
		result[i] = string(lang)
	}
	return result
}

// ParseLanguage parses a string into a Language
func ParseLanguage(s string) (Language, error) {
	lang := Language(s)
	if !lang.IsValid() {
		return "", fmt.Errorf("invalid language: %s (valid options: go, dotnet, nodejs)", s)
	}
	return lang, nil
}

// DefaultFileExtension returns the default file extension for the language
func (l Language) DefaultFileExtension() string {
	switch l {
	case LanguageGo:
		return ".go"
	case LanguageDotNet:
		return ".cs"
	case LanguageNodeJS:
		return ".ts"
	default:
		return ""
	}
}
