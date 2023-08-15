package grest

import "strings"

// String represents utility functions for string manipulation.
type String struct{}

// AddSlashes adds slashes before characters that need to be escaped.
// These characters are:
//   - single quote (')
//   - double quote (")
//   - and backslash (\)
func (String) AddSlashes(str string) string {
	b := strings.Builder{}
	for _, r := range []rune(str) {
		switch r {
		case []rune{'\\'}[0], []rune{'"'}[0], []rune{'\''}[0]:
			b.WriteRune([]rune{'\\'}[0])
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// StripSlashes un-quotes a quoted string.
func (String) StripSlashes(str string) string {
	b := strings.Builder{}
	strRune := []rune(str)
	for i := 0; i < len(strRune); i++ {
		if strRune[i] == []rune{'\\'}[0] {
			i++
		}
		b.WriteRune(strRune[i])
	}
	return b.String()
}

// CamelCase converts a string to camelCase.
func (s String) CamelCase(str string, startWithUpper ...bool) string {
	b := strings.Builder{}
	strRune := []rune(str)
	lenStr := len(strRune)
	for i := 0; i < lenStr; i++ {
		if i == 0 {
			i = s.FirstAlphaRuneIndex(strRune, i)
			if i < lenStr {
				if len(startWithUpper) > 0 && startWithUpper[0] {
					b.WriteRune(s.ToUpperAlphaRune(strRune[i]))
				} else {
					b.WriteRune(s.ToLowerAlphaRune(strRune[i]))
				}
			}
		} else {
			if !s.IsAlphaNumericRune(strRune[i]) {
				i = s.FirstAlphaNumericRuneIndex(strRune, i)
				if i < lenStr {
					b.WriteRune(s.ToUpperAlphaRune(strRune[i]))
				}
			} else {
				b.WriteRune(strRune[i])
			}
		}
	}
	return b.String()
}

// PascalCase converts a string to PascalCase.
func (s String) PascalCase(str string) string {
	return s.CamelCase(str, true)
}

// SpecialCase converts a string to snake_case, kebab-case, or other formats based on the delimiter.
func (s String) SpecialCase(str string, delimiter rune) string {
	b := strings.Builder{}
	strRune := []rune(str)
	lenStr := len(strRune)
	for i := 0; i < lenStr; i++ {
		if i == 0 {
			i = s.FirstAlphaRuneIndex(strRune, i)
			if i < lenStr {
				b.WriteRune(s.ToLowerAlphaRune(strRune[i]))
			}
		} else {
			if !s.IsLowerAlphaRune(strRune[i]) && !s.IsNumericRune(strRune[i]) {
				i = s.FirstAlphaNumericRuneIndex(strRune, i)
				if i < lenStr {
					b.WriteRune(delimiter)
					b.WriteRune(s.ToLowerAlphaRune(strRune[i]))
				}
			} else {
				b.WriteRune(strRune[i])
			}
		}
	}
	return b.String()
}

// SnakeCase converts a string to snake_case.
func (s String) SnakeCase(str string) string {
	return s.SpecialCase(str, '_')
}

// KebabCase converts a string to kebab-case.
func (s String) KebabCase(str string) string {
	return s.SpecialCase(str, '-')
}

// IsLowerAlphaRune checks if a rune is a lowercase alphabetic character.
func (String) IsLowerAlphaRune(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// IsUpperAlphaRune checks if a rune is an uppercase alphabetic character.
func (String) IsUpperAlphaRune(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// IsNumericRune checks if a rune is a numeric character.
func (String) IsNumericRune(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsAlphaNumericRune checks if a rune is an alphanumeric character.
func (s String) IsAlphaNumericRune(r rune) bool {
	return s.IsLowerAlphaRune(r) || s.IsUpperAlphaRune(r) || s.IsNumericRune(r)
}

// ToLowerAlphaRune converts an uppercase alphabetic rune to lowercase.
func (s String) ToLowerAlphaRune(r rune) rune {
	if s.IsUpperAlphaRune(r) {
		return r + 'a' - 'A'
	}
	return r
}

// ToUpperAlphaRune converts a lowercase alphabetic rune to uppercase.
func (s String) ToUpperAlphaRune(r rune) rune {
	if s.IsLowerAlphaRune(r) {
		return r - 'a' + 'A'
	}
	return r
}

// FirstAlphaRuneIndex finds the index of the first alphabetic rune.
func (s String) FirstAlphaRuneIndex(sr []rune, start int) int {
	for i, r := range sr {
		if i >= start && (s.IsLowerAlphaRune(r) || s.IsUpperAlphaRune(r)) {
			return i
		}
	}
	return len(sr)
}

// FirstAlphaNumericRuneIndex finds the index of the first alphanumeric rune.
func (s String) FirstAlphaNumericRuneIndex(sr []rune, start int) int {
	for i, r := range sr {
		if i >= start && (s.IsLowerAlphaRune(r) || s.IsUpperAlphaRune(r) || s.IsNumericRune(r)) {
			return i
		}
	}
	return len(sr)
}

// GetVars extracts variables from a string using specified before and after delimiters.
func (String) GetVars(str, before, after string) []string {
	vars := []string{}
	temp := strings.Split(str, before)
	for _, v := range temp {
		if strings.Contains(v, after) {
			vars = append(vars, strings.Split(v, after)[0])
		}
	}
	return vars
}
