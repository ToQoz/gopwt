package translatedassert

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type diffType int

func (t diffType) String() string {
	switch t {
	case diffTypeLine:
		return "line"
	case diffTypeChar:
		return "char"
	}
	panic("unexpected value")
}

const (
	diffTypeChar diffType = iota
	diffTypeLine
)

func diff(pa, pb posValuePair) (string, diffType) {
	a := formatValueForDiff(pa.Value)
	b := formatValueForDiff(pb.Value)

	dtype := diffTypeChar

	diffs := charaDiff(a, b)
	eql := "%s"
	ins := "{+%s+}"
	del := "[-%s-]"

	total := map[diffmatchpatch.Operation]int{
		diffmatchpatch.DiffEqual:  0,
		diffmatchpatch.DiffInsert: 0,
		diffmatchpatch.DiffDelete: 0,
	}

	for _, d := range diffs {
		total[d.Type] += len(d.Text)
		if d.Type == diffmatchpatch.DiffInsert || d.Type == diffmatchpatch.DiffDelete {
			// if diffs have a multiline chara-diff, use line-diff instead of it
			if strings.Contains(strings.TrimSuffix(d.Text, "\n"), "\n") {
				dtype = diffTypeLine
				break
			}
		}
	}
	// changes are less than 10%, use line-diff
	changes := total[diffmatchpatch.DiffInsert] + total[diffmatchpatch.DiffDelete]
	if float64(changes)/float64(changes+total[diffmatchpatch.DiffEqual]) > 0.2 {
		dtype = diffTypeLine
	}

	ret := ""
	ret += fmt.Sprintf("--- [%T] %s\n", pa.Value, pa.OriginalExpr)
	ret += fmt.Sprintf("+++ [%T] %s\n", pb.Value, pb.OriginalExpr)
	ret += "@@ -1," + strconv.Itoa(strings.Count(a, "\n")+1) + " +1," + strconv.Itoa(strings.Count(b, "\n")+1) + "@@\n"
	if dtype == diffTypeLine {
		diffs = lineDiff(a, b)
		eql = " %s"
		ins = "+%s"
		del = "-%s"
	}
	for _, d := range diffs {
		var format string
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			format = green(ins)
		case diffmatchpatch.DiffDelete:
			format = red(del)
		case diffmatchpatch.DiffEqual:
			format = eql
		}
		ret += fmt.Sprintf(format, d.Text)
	}

	return ret, dtype
}

func green(s string) string {
	if !stdoutIsatty {
		return s
	}

	return "\x1b[32m" + s + "\x1b[39m"
}

func red(s string) string {
	if !stdoutIsatty {
		return s
	}

	return "\x1b[31m" + s + "\x1b[39m"
}

func charaDiff(a, b string) []diffmatchpatch.Diff {
	dmp := diffmatchpatch.New()
	return dmp.DiffMain(a, b, true)
}
func lineDiff(a, b string) []diffmatchpatch.Diff {
	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(a, b)
	diffs := dmp.DiffMain(a, b, false)
	lineDiffs := dmp.DiffCharsToLines(diffs, c)

	result := []diffmatchpatch.Diff{}
	for _, ld := range lineDiffs {
		for _, l := range strings.Split(strings.TrimSuffix(ld.Text, "\n"), "\n") {
			result = append(result, diffmatchpatch.Diff{Text: l + "\n", Type: ld.Type})
		}
	}
	return result
}
