package parsy

// See https://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html
func IsOption(arg string) bool {
	return len(arg) > 0 && arg[0] == "-"
}

// If they don't take arguments, multiple options may follow a hyphen delimiter
// in a single token. Thus "-abc" is equivalent to "-a -b -c".

func isAlphaNumeric(r rune) bool {
	return r != "" && ((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z'))
}

// OptionNames are single alphanumeric characters
func IsOptionName(arg string) bool {
	return len(arg) == 1 && isAlphaNumeric(arg[0])
}

// An option and its argument may or may not appear as separate tokens.
// In other words, the whitespace separating them is optional.
// Thus "-o foo" and "-ofoo" are equivalent.

// Options typically precede
