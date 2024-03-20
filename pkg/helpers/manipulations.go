package helpers

import (
	"fmt"
	"strings"
)

// RenameDuplicates takes an input slice of strings and renames any duplicate headers
// by appending a count to them. It returns the modified input slice.
//
// Each header in the input slice is checked against a map called counts. The map stores
// the count of each header occurrence. If a header occurs more than once, its count is
// incremented and the header is renamed by appending "_<count>" to it.
//
// After the renaming is done, the counts map is iterated to print a message for each header
// that had duplicates.
//
// Example usage:
//
//	headers := []string{"Name", "Age", "Name", "City", "Age"}
//	modifiedHeaders := RenameDuplicates(headers)
//
// Output:
//
//	Header 'Name' was present 2 times
//	Header 'Age' was present 2 times
//	Header 'Name_2' was present 1 times
//	Header 'City' was present 1 times
//
//	The modifiedHeaders slice will be:
//	[]string{"Name", "Age", "Name_2", "City", "Age_2"}
func RenameDuplicates(input []string) []string {
	counts := make(map[string]int)

	for i, header := range input {
		counts[header]++
		if counts[header] > 1 {
			input[i] = fmt.Sprintf("%s_%d", header, counts[header])
		}
	}
	for header, count := range counts {
		if count > 1 {
			fmt.Printf("Header '%s' was present %d times\n", header, count)
		}
	}
	return input
}

// FixXMLTags takes a string `tag` as input and removes any invalid XML characters from it.
// It returns the modified string with the cleaned tag.
// The function first initializes a slice `invalidXmlChars` with a list of invalid XML characters.
// These characters are identified as parentheses, angle brackets, slashes, backslashes,
// question marks, exclamation marks, double and single quotation marks, at signs, hash signs, dollar signs,
// percent signs, caret symbols, ampersands, asterisks, plus signs, equal signs, tilde, backticks,
// vertical bars, square brackets, curly braces, semicolons, colons, commas, and periods.
// The function then iterates over each character in the `invalidXmlChars` slice.
// For each character, it removes all occurrences of that character in the `tag` string
// using the `ReplaceAll` function from the `strings` package,
// and assigns the result back to the `cleanTag` variable.
// Finally, the function returns the `cleanTag` string, which contains the modified tag
// with all invalid XML characters removed.
// Example usage:
//
//	tag := "<Hello World!>"
//	cleanTag := FixXMLTags(tag)
//	fmt.Println(cleanTag)
//	// Output: "Hello World"
func FixXMLTags(tag string) string {
	invalidXmlChars := []rune{
		'(', ')', '<', '>', '/', '\\',
		'?', '!', '"', '\'', '@', '#', '$',
		'%', '^', '&', '*', '+', '=', '~',
		'`', '|', '[', ']', '{', '}', ';',
		':', ',', '.',
	}
	// Replace invalid characters
	cleanTag := tag
	for _, char := range invalidXmlChars {
		cleanTag = strings.ReplaceAll(cleanTag, string(char), "")
	}
	return cleanTag
}
