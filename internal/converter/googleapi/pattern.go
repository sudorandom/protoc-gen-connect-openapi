package googleapi

import (
	"fmt"
	"strings"
	"unicode"
)

// Grammar:
// Template = "/" Segments [ Verb ] ;
// Segments = Segment { "/" Segment } ;
// Segment  = "*" | "**" | LITERAL | Variable ;
// Variable = "{" FieldPath [ "=" Segments ] "}" ;
// FieldPath = IDENT { "." IDENT } ;
// Verb     = ":" LITERAL ;

// Token represents a lexical token.
type Token struct {
	Type  TokenType
	Value string
}

// TokenType represents the possible token types.
type TokenType string

const (
	TokenSlash    TokenType = "SLASH"
	TokenColon    TokenType = "COLON"
	TokenLiteral  TokenType = "LITERAL"
	TokenVariable TokenType = "VARIABLE"
	TokenIdent    TokenType = "IDENT"
	TokenEOF      TokenType = "EOF"
)

// RunPathPatternLexer takes an input string and returns a stream of tokens.
func RunPathPatternLexer(input string) ([]Token, error) {
	var tokens []Token
	runes := []rune(input)
	length := len(runes)
	pos := 0

	for pos < length {
		switch {
		case unicode.IsSpace(runes[pos]):
			pos++
		case runes[pos] == '/':
			tokens = append(tokens, Token{Type: TokenSlash, Value: string(runes[pos])})
			pos++
		case runes[pos] == ':':
			tokens = append(tokens, Token{Type: TokenColon, Value: string(runes[pos])})
			pos++
		case isWordChar(runes[pos]):
			word := getWord(runes[pos:])
			word = strings.Split(word, ":")[0]
			if isLiteral(word) {
				tokens = append(tokens, Token{Type: TokenLiteral, Value: word})
			} else if isVariable(word) {
				tokens = append(tokens, Token{Type: TokenVariable, Value: word[1 : len(word)-1]})
			} else if isIdentStart(rune(word[0])) {
				tokens = append(tokens, Token{Type: TokenIdent, Value: word})
			} else {
				return nil, fmt.Errorf("unrecognized word at position: %d: %s", pos, word)
			}
			pos += len(word)
		default:
			// Handle error for unrecognized character
			pos++
		}
	}

	tokens = append(tokens, Token{Type: TokenEOF, Value: ""})
	return tokens, nil
}

// Helper function to extract a word starting from a given position in the input.
func getWord(input []rune) string {
	var word []rune
	inBrackets := false
	for i, char := range input {
		if !inBrackets && char == '/' {
			break
		}
		if char == '{' && inBrackets {
			// Multiple bracket sections isn't valid, so start parsing the word here
			break
		}
		if char == '{' {
			inBrackets = true
		}
		if char == '}' {
			inBrackets = false
		}
		word = append(word, char)
		if i == len(input)-1 {
			break
		}
	}
	return string(word)
}

func isWordChar(char rune) bool {
	return char != '/'
}

// Helper functions to check the type of word.
func isLiteral(word string) bool {
	literalOptions := []string{"*", "**"}
	for _, option := range literalOptions {
		if word == option {
			return true
		}
	}
	return false
}

func isVariable(word string) bool {
	if len(word) < 3 {
		return false
	}
	return word[0] == '{' && word[len(word)-1] == '}'
}

func isIdentStart(char rune) bool {
	return ((char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_' ||
		char == '.')
}
