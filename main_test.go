package main

import "testing"

func Test_cleanProfanity(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		text string
		want string
	}{
		// TODO: Add test cases.
		{
			name: "text with no profacnity returns original text",
			text: "My momma raised me not to be a cusser.",
			want: "My momma raised me not to be a cusser.",
		},
		{
			name: "text with sharbert returns cleaned text",
			text: "My momma raised me not to be a sharbert cusser.",
			want: "My momma raised me not to be a **** cusser.",
		},
		{
			name: "text with kerfuffle returns cleaned text",
			text: "My momma raised me not to be a kerfuffle cusser.",
			want: "My momma raised me not to be a **** cusser.",
		},
		{
			name: "text with fornax returns cleaned text",
			text: "My momma raised me not to be a fornax cusser.",
			want: "My momma raised me not to be a **** cusser.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanProfanity(tt.text)
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("cleanProfanity() = %v, want %v", got, tt.want)
			}
		})
	}
}
