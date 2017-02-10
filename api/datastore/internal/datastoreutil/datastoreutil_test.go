package datastoreutil

import (
	"testing"
)

func TestStripParamNames(t *testing.T) {
	for _, testCase := range []struct{ input, expected string }{
		{`/blogs`, `/blogs`},
		{`/blogs/:blog_id`, `/blogs/:`},
		{`/blogs/:blog_id/comments`, `/blogs/:/comments`},
		{`/blogs/:blog_id/comments/:comment_id`, `/blogs/:/comments/:`},
		{`/test/*catchAll`, `/test/*`},
	} {
		got := StripParamNames(testCase.input)
		if testCase.expected != got {
			t.Errorf("for %q: expected %q but got %q", testCase.input, testCase.expected, got)
		}
	}
}

func TestSqlLikeToRegExp(t *testing.T) {
	for _, testCase := range []struct{ input, expected string }{
		{`ab%`, `^ab.*?$`},
		{`%ab`, `^.*?ab$`},
		{`a%b`, `^a.*?b$`},
	} {
		got := SqlLikeToRegExp(testCase.input)
		if testCase.expected != got {
			t.Errorf("for %q: expected %q but got %q", testCase.input, testCase.expected, got)
		}
	}
}
