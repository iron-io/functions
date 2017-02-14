package datastoreutil

import (
	"testing"
	"reflect"
)

func TestPathParts(t *testing.T) {
	for _, test := range []struct{
		path string
		expectedParts []string
		expectedSlash bool
	} {
		{`/`, nil, false},
		{`/test`, []string{"test"}, false},
		{`/test/`, []string{"test"}, true},
		{`/test/test`, []string{"test","test"}, false},
		{`/test/test/`, []string{"test","test"}, true},
	} {
		gotParts, gotSlash := SplitPath(test.path)
		if !reflect.DeepEqual(gotParts, test.expectedParts) {
			t.Errorf("for %q: expected parts %v but got %v", test.path, test.expectedParts, gotParts)
		}
		if gotSlash != test.expectedSlash {
			t.Errorf("for %q: expected slash %t but got %t", test.path, test.expectedSlash, gotSlash)
		}
	}
}

func TestStripParamNames(t *testing.T) {
	for _, testCase := range []struct{ input, expected string }{
		{`/blogs`, `/blogs`},
		{`/blogs/`, `/blogs/`},
		{`/blogs/:blog_id`, `/blogs/:`},
		{`/blogs/:blog_id/`, `/blogs/:/`},
		{`/blogs/:blog_id/comments`, `/blogs/:/comments`},
		{`/blogs/:blog_id/comments/:comment_id`, `/blogs/:/comments/:`},
		{`/blogs/:blog_id/comments/:comment_id/*suffix`, `/blogs/:/comments/:/*`},
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
