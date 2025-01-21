package utils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

type StructPath []StructPathSegment

func NewStructPath(label string) (out StructPath) {
	parts := strings.Split(label, ".")
	for _, part := range parts {
		out = append(out, StructPathSegment(part))
	}
	return out
}

func (l StructPath) String() string {
	out := make([]string, len(l))
	for idx, key := range l {
		out[idx] = strcase.ToSnake(strings.ToLower(key.String()))
	}
	return strings.Join(out, ".")
}

func (l StructPath) Append(in string) StructPath {
	l = append(l, StructPathSegment(in))
	return l
}

func (l StructPath) isLast(offset uint64) bool {
	return int(offset) == len(l)-1
}

type StructPathSegment string

func (l StructPathSegment) IsKey() bool {
	return !l.IsArrayIter()
}

func (l StructPathSegment) String() string {
	return string(l)
}

func (l StructPathSegment) GetKey() string {
	return l.String()
}

var decimal = regexp.MustCompile(`^(\d+|\[])$`)

func (l StructPathSegment) IsArrayIter() bool {
	return decimal.MatchString(l.String())
}

func (l StructPathSegment) GetIndex() uint64 {
	if index, err := strconv.ParseUint(l.String(), 10, 64); err == nil {
		return index
	}
	return 0
}
