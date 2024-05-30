package util

import (
	"strconv"
	"testing"
)

var specialCharMap = make(map[string]struct{})

func TestGeneratePassword(t *testing.T) {
	for i := range specialChars {
		specialCharMap[string(specialChars[i])] = struct{}{}
	}

	for i := 0; i < 1000; i++ {
		p, err := GeneratePassword()
		if err != nil {
			t.Fatalf("error generating password: %v", err)
		}
		counts := make([]int, 3)
		for j := range p {
			counts[0] += 1
			c := string(p[j])
			_, ok := specialCharMap[c]
			if ok {
				counts[1] += 1
				continue
			}
			_, err := strconv.Atoi(c)
			if err == nil {
				counts[2] += 1
			}
		}
		if counts[0] != passwordLength {
			t.Fatalf("password length incorrect")
		}
		if counts[1] < numSpecialChars {
			t.Fatalf("min number of special chars not generated")
		}
		if counts[2] < numDigits {
			t.Fatalf("min number of digits not generated")
		}
	}
}
