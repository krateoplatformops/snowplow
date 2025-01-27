package tmpl

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"
)

func dataSource() (any, error) {
	const sample = `
	{
		"metadata": {
			"labels": {
		 		"krateo.io/composition-id": "XXXXXX" 
			}
		},
		"__internal_ep_ref_name": "tizio-clientconfig",
		"__internal_ep_ref_namespace": "demo-system",
		"firstName": "Charles",
		"lastName": "Doe",
		"age": 41,
		"location": {
		  "city": "San Fracisco",
		  "postalCode": "94103"
		},
		"hobbies": [
		  "chess",
		  "netflix"
		],
		"id": 1
	  }`

	var res any
	err := json.Unmarshal([]byte(sample), &res)
	return res, err
}

func TestRegexPatternBuild(t *testing.T) {
	leftDelim, rightDelim := "${", "}"
	pattern := fmt.Sprintf("^%s\\s+(.*)%s",
		regexp.QuoteMeta(leftDelim),
		regexp.QuoteMeta(rightDelim))

	fmt.Println(pattern)
}

func TestJQTemplate(t *testing.T) {
	test := []struct {
		input string
		want  string
	}{
		{`${ .age }`, "41"},
		{` .age }}`, ` .age }}`},
		{`${ .location.city }`, "San Fracisco"},
		{"hello world", "hello world"},
		{`${ .hobbies | join(",") }`, "chess,netflix"},
		{`${ .id }`, "1"},
		{`${ "/todos/" + (.id|tostring) }`, "/todos/1"},
		{`${ "/todos/" + (.id|tostring) +  "/comments" }`, "/todos/1/comments"},
		{`${ .__internal_ep_ref_name }`, "tizio-clientconfig"},
		{`${ .__internal_ep_ref_namespace }`, "demo-system"},
	}

	ds, err := dataSource()
	if err != nil {
		t.Fatal(err)
	}

	tpl, err := New("${", "}")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range test {
		got, err := tpl.Execute(tc.input, ds)
		if err != nil {
			t.Fatal(err)
		}

		if got != tc.want {
			t.Fatalf("got: %s, want: %s\n", got, tc.want)
		}
	}
}

func TestAcceptQuery(t *testing.T) {
	test := []struct {
		input string
		want  string
		ok    bool
	}{
		{`${ .age }`, `.age`, true},
		{` .age }}`, ` .age }}`, false},
		{`${ .location.city }`, `.location.city`, true},
		{`hello world`, `hello world`, false},
		{`${ .hobbies | join(",") }`, `.hobbies | join(",")`, true},
	}

	tpl, err := New("${", "}")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range test {
		got, ok := tpl.ParseQuery(tc.input)
		if got != tc.want {
			t.Fatalf("got: %s, want: %s\n", got, tc.want)
		}
		if ok != tc.ok {
			t.Fatalf("got: %s, want: %s\n", got, tc.want)
		}
	}
}

func Example_JQ_Execute() {
	sample := `
[
	{
		"color": "red",
		"value": "#f00"
	},
	{
		"color": "green",
		"value": "#0f0"
	},
	{
		"color": "blue",
		"value": "#00f"
	},
	{
		"color": "cyan",
		"value": "#0ff"
	},
	{
		"color": "magenta",
		"value": "#f0f"
	},
	{
		"color": "yellow",
		"value": "#ff0"
	},
	{
		"color": "black",
		"value": "#000"
	}
]`

	var data any
	err := json.Unmarshal([]byte(sample), &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	tpl, err := New("${", "}")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	got, err := tpl.Execute("${ .[2:4] }", data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Printf("%s\n", got)

	// Output:
	// [{"color":"blue","value":"#00f"},{"color":"cyan","value":"#0ff"}]
}
