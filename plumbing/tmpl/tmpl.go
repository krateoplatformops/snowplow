package tmpl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/itchyny/gojq"
)

func New(leftDelim, rightDelim string) (*JQ, error) {
	pattern := fmt.Sprintf("^%s\\s+(.*)%s",
		regexp.QuoteMeta(leftDelim),
		regexp.QuoteMeta(rightDelim))

	re, err := regexp.Compile(pattern) // `^\{\{\s+(.*)\}\}$`)
	if err != nil {
		return nil, err
	}

	return &JQ{
		re:      re,
		unquote: true,
	}, nil
}

type JQ struct {
	re      *regexp.Regexp
	unquote bool
}

func (t *JQ) Execute(q string, data any) (string, error) {
	q, ok := t.ParseQuery(q)
	if !ok {
		return q, nil
	}

	enc := newEncoder(false, 0)

	query, err := gojq.Parse(q)
	if err != nil {
		return "", err
	}

	iter := query.Run(data) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return "", err
		}
		if err := enc.encode(v); err != nil {
			return "", err
		}
	}

	res := enc.w.String()
	if t.unquote {
		unq, err := strconv.Unquote(res)
		if err == nil {
			res = unq
		}
	}

	return res, nil
}

func (t *JQ) ParseQuery(q string) (string, bool) {
	if !t.re.MatchString(q) {
		return q, false
	}

	res := t.re.FindAllStringSubmatch(q, -1)
	if len(res) == 0 || len(res[0]) == 0 {
		return q, false
	}

	return strings.TrimSpace(res[0][1]), true
}
