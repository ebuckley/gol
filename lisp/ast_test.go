package lisp

import (
	"testing"
)

func Test_readFromTokens(t *testing.T) {
	lexed := []Token{
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
	}
	ast, toks, err := recursiveReadAST(lexed[0], lexed[1:])
	if err != nil {
		t.Fatalf("error reading tokens %v", err)
		return
	}
	if len(toks) != 1 {
		t.Fatalf("Should have just one remaining token for the eof: %v", toks)
	}

	_, err = Eval(ast, DefaultScope())
	if err != nil {
		t.Fatal("should not error but we got", err.Error())
	}
	//type args struct {
	//	ts []Token
	//}
	//tests := []struct {
	//	name    string
	//	args    args
	//	want    Node
	//	wantErr bool
	//}{
	//	,
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		got, err := readFromTokens(tt.args.ts)
	//		if (err != nil) != tt.wantErr {
	//			t.Errorf("readFromTokens() error = %v, wantErr %v", err, tt.wantErr)
	//			return
	//		}
	//		if !reflect.DeepEqual(got, tt.want) {
	//			t.Errorf("readFromTokens() got = %v, want %v", got, tt.want)
	//		}
	//	})
	//}
}
