package main

import (
	"fmt"
	"testing"
)

func TestMust(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("must panic if err is given")
			}
		}()

		must(fmt.Errorf("error"))
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("must not panic if err is not given")
			}
		}()

		must(nil)
	}()
}
