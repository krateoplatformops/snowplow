//go:build unit
// +build unit

package jqutil

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		unquote bool
		data    any
		want    string
		wantErr bool
	}{
		{
			name:    "Extract name from object",
			query:   ".name",
			unquote: false,
			data:    map[string]any{"name": "Alice"},
			want:    `"Alice"`,
			wantErr: false,
		},
		{
			name:    "Extract number",
			query:   ".age",
			unquote: false,
			data:    map[string]any{"age": 25},
			want:    "25",
			wantErr: false,
		},
		{
			name:    "Extract string with unquote",
			query:   ".city",
			unquote: true,
			data:    map[string]any{"city": "Paris"},
			want:    "Paris",
			wantErr: false,
		},
		{
			name:    "Extract from simple list",
			query:   "[.items[] | .name] | sort",
			unquote: false,
			data: map[string]any{
				"items": []any{
					map[string]any{"name": "Alice"},
					map[string]any{"name": "Bob"},
					map[string]any{"name": "John"},
					map[string]any{"name": "Sammy"},
					map[string]any{"name": "Carol"},
				},
			},
			want:    `["Alice","Bob","Carol","John","Sammy"]`,
			wantErr: false,
		},
		{
			name:    "Extract from complex list",
			query:   "[.items | sort_by(.age)[] | {name, job}]",
			unquote: false,
			data: map[string]any{
				"items": []any{
					map[string]any{"name": "Alice", "job": "Engineer", "age": 30},
					map[string]any{"name": "Bob", "job": "Designer", "age": 25},
					map[string]any{"name": "John", "job": "Manager", "age": 40},
					map[string]any{"name": "Sammy", "job": "Developer", "age": 28},
					map[string]any{"name": "Carol", "job": "HR", "age": 35},
				},
			},
			want:    `[{"job":"Designer","name":"Bob"},{"job":"Developer","name":"Sammy"},{"job":"Engineer","name":"Alice"},{"job":"HR","name":"Carol"},{"job":"Manager","name":"John"}]`,
			wantErr: false,
		},
		{
			name:    "Iterate list ",
			query:   `[.items[] | "/todos?user=" + .name]`,
			unquote: false,
			data: map[string]any{
				"items": []any{
					map[string]any{"name": "Alice", "job": "Engineer", "age": 30},
					map[string]any{"name": "Bob", "job": "Designer", "age": 25},
					map[string]any{"name": "John", "job": "Manager", "age": 40},
					map[string]any{"name": "Sammy", "job": "Developer", "age": 28},
					map[string]any{"name": "Carol", "job": "HR", "age": 35},
				},
			},
			want:    `["/todos?user=Alice","/todos?user=Bob","/todos?user=John","/todos?user=Sammy","/todos?user=Carol"]`,
			wantErr: false,
		},
		{
			name:    "Invalid query",
			query:   ".invalid(",
			unquote: false,
			data:    map[string]any{"key": "value"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Create a new JSON",
			query:   "{ compositionID : .item.uid }",
			unquote: false,
			data: map[string]any{
				"item": map[string]any{"uid": "AA-BB-CC", "name": "Alice"},
			},
			want:    `{"compositionID":"AA-BB-CC"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Eval(context.TODO(), EvalOptions{
				Query:   tt.query,
				Unquote: tt.unquote,
				Data:    tt.data,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Eval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtract(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		unquote bool
		data    any
		want    map[string]any
		wantErr bool
	}{
		{
			name:    "Extract valid array of objects",
			query:   "[.employees[] | {name}]",
			unquote: false,
			data: map[string]any{
				"employees": []any{
					map[string]any{"name": "Bob", "job": "Designer"},
					map[string]any{"name": "Sammy", "job": "Developer"},
					map[string]any{"name": "Alice", "job": "Engineer"},
				},
			},
			want: map[string]any{
				"items": []any{
					map[string]any{"name": "Bob"},
					map[string]any{"name": "Sammy"},
					map[string]any{"name": "Alice"},
				},
			},
			wantErr: false,
		},
		{
			name:    "Extract empty array",
			query:   "[.employees[] | .name]",
			unquote: false,
			data: map[string]any{
				"employees": []any{},
			},
			want: map[string]any{
				"items": []any{},
			},
			wantErr: false,
		},
		{
			name:    "Extract string array",
			query:   `[.employees[] | "/users?name=" + (.name)]`,
			unquote: false,
			data: map[string]any{
				"employees": []any{
					map[string]any{"name": "Bob", "job": "Designer"},
					map[string]any{"name": "Sammy", "job": "Developer"},
					map[string]any{"name": "Alice", "job": "Engineer"},
				},
			},
			want: map[string]any{
				"items": []any{"/users?name=Bob", "/users?name=Sammy", "/users?name=Alice"},
			},
			wantErr: false,
		},
		{
			name:    "Invalid query that doesn't return an array",
			query:   ".employees.name", // non restituisce un array
			unquote: false,
			data: map[string]any{
				"employees": []any{
					map[string]any{"name": "Bob", "job": "Designer"},
				},
			},
			want:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Extract(context.TODO(), EvalOptions{
				Query:   tt.query,
				Unquote: tt.unquote,
				Data:    tt.data,
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestForEach(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		data    any
		action  func(any) error
		wantErr bool
	}{
		{
			name: "Process valid JSON array",
			data: map[string]any{
				"items": []any{
					map[string]any{"name": "Alice", "job": "Engineer", "age": 30},
					map[string]any{"name": "Bob", "job": "Designer", "age": 25},
					map[string]any{"name": "John", "job": "Manager", "age": 40},
					map[string]any{"name": "Sammy", "job": "Developer", "age": 28},
					map[string]any{"name": "Carol", "job": "HR", "age": 35},
				},
			},
			query: "[.items[]] | sort_by(.age)",
			action: func(el any) error {
				obj, _ := el.(map[string]any)
				fmt.Println("Processing:", obj["name"])
				return nil
			},
			wantErr: false,
		},
		{
			name: "Empty JSON array",
			data: map[string]any{
				"items": []any{},
			},
			query: "[]",
			action: func(el any) error {
				return errors.New("this should not be called")
			},
			wantErr: false,
		},
		{
			name: "Query does not return array",
			data: map[string]any{
				"items": []any{
					map[string]any{"name": "Alice", "job": "Engineer", "age": 30},
					map[string]any{"name": "Bob", "job": "Designer", "age": 25},
					map[string]any{"name": "John", "job": "Manager", "age": 40},
					map[string]any{"name": "Sammy", "job": "Developer", "age": 28},
					map[string]any{"name": "Carol", "job": "HR", "age": 35},
				},
			},
			query: "invalid_query",
			action: func(el any) error {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := EvalOptions{Query: tt.query, Data: tt.data}
			ctx := context.Background()
			err := ForEach(ctx, opts, tt.action)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMaybeQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		found    bool
	}{
		{
			name:     "Basic extraction",
			input:    "Hello ${name}",
			expected: "name",
			found:    true,
		},
		{
			name:     "Extraction with spaces",
			input:    "Hello ${  username   }",
			expected: "username",
			found:    true,
		},
		{
			name:     "No placeholder",
			input:    "Hello world",
			expected: "Hello world",
			found:    false,
		},
		{
			name:     "Unclosed placeholder",
			input:    "Hello ${name",
			expected: "Hello ${name",
			found:    false,
		},
		{
			name:     "Multiple placeholders (only first is extracted)",
			input:    "Hello ${name}, welcome to ${city}",
			expected: "name",
			found:    true,
		},
		{
			name:     "Only placeholder ${}",
			input:    "${}",
			expected: "",
			found:    true,
		},
		{
			name:     "Placeholder at the end",
			input:    "Welcome to ${location}",
			expected: "location",
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := MaybeQuery(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.found, found)
		})
	}
}
