package lisp

import (
	"reflect"
	"testing"
)


func TestLexer_NextToken(t *testing.T) {

	tests := []struct {
		name   string
		input string
		want   []Token
	}{
		{
			name: "Simple case",
			input: "(+ 1 2)",
			want: []Token{
				{
					Type:    LPAREN,
					Literal: "",
				},
				{
					Type: SYMBOL,
					Literal: "+",
				},
				{
					Type: SYMBOL,
					Literal: "1",
				},
				{
					Type: SYMBOL,
					Literal: "2",
				},
				{
					Type:    RPAREN,
					Literal: "",
				},
				{
					Type:    EOF,
					Literal: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			for i, wantedToken := range tt.want {
				if got := l.NextToken(); !reflect.DeepEqual(got,wantedToken) {
					t.Errorf("Test [%s] i[%d] NextToken() = %v, want %v", tt.name, i, got, wantedToken)
				}
			}
		})
	}
}
