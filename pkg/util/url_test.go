package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestURLParse(t *testing.T) {

	type expect struct {
		cloneUrl  string
		path      string
		ref       string
		submodule bool
		timeout   time.Duration
		err       bool
	}

	type testCase struct {
		expect expect
		input  string
	}

	cases := []testCase{
		{
			input: "https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?timeout=120&ref=v3.3.1",
			expect: expect{
				cloneUrl:  "https://github.com/kubernetes-sigs/kustomize",
				path:      "examples/multibases/dev",
				ref:       "v3.3.1",
				submodule: true,
				timeout:   120 * time.Second,
			},
		},
		{
			input: "git@github.com:owner/repo//examples?timeout=120&version=v3.3.1",
			expect: expect{
				cloneUrl:  "git@github.com:owner/repo",
				path:      "examples",
				ref:       "v3.3.1",
				submodule: true,
				timeout:   120 * time.Second,
			},
		},
		{
			input: "https://   /(@kubernetes-sigs/kustomize//examples/multibases/dev/?timeout=120&ref=v3.3.1",
			expect: expect{
				err: true,
			},
		},
		{
			input: "https://my.github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?version=v3.3.1&submodules=false&timeout=1s",
			expect: expect{
				cloneUrl:  "https://my.github.com/kubernetes-sigs/kustomize",
				path:      "examples/multibases/dev",
				ref:       "v3.3.1",
				submodule: false,
				timeout:   1 * time.Second,
			},
		},
	}

	for i := range cases {
		c := cases[i]

		r, err := NewKustomizeRemote(c.input)
		if err != nil {
			if !c.expect.err {
				assert.Fail(t, err.Error())
			} else {
				continue
			}
		}
		assert.Equal(t, c.expect.path, r.Path())
		assert.Equal(t, c.expect.cloneUrl, r.CloneUrl())
		assert.Equal(t, c.expect.timeout, r.Timeout)
		assert.Equal(t, c.expect.ref, r.Ref)
		assert.Equal(t, c.expect.submodule, r.Submodules)
	}
}
