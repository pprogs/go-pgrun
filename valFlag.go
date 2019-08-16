package main

import (
	"errors"
	"fmt"
	"strings"
)

type valFlag []string

//String f
func (v *valFlag) String() string {
	return strings.Join(*v, ",")
}

func (v *valFlag) Replacer() *strings.Replacer {
	return strings.NewReplacer((*v)...)
}

//Add f
func (v *valFlag) Add(name string, value string) error {

	if len(name) == 0 || len(value) == 0 {
		return errors.New("name or value is empty")
	}
	n := fmt.Sprintf("##%s##", name)

	for i := 0; i < len(*v); i = i + 2 {
		if (*v)[i] == n {
			(*v)[i+1] = value
			return nil
		}
	}

	*v = append(*v, n, value)

	return nil
}

//Set f
func (v *valFlag) Set(value string) error {

	s := strings.Split(value, ",")
	if len(s) != 2 {
		return errors.New("error parsing values")
	}
	return v.Add(s[0], s[1])
}
