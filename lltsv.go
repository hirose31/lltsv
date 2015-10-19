package main

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/andrew-d/go-termutil"
	"github.com/mgutz/ansi"
)

type tFuncAppend func([]string, string, string) []string

type Lltsv struct {
	keys       []string
	no_key     bool
	funcAppend tFuncAppend
}

func newLltsv(keys []string, no_key bool) *Lltsv {
	return &Lltsv{
		keys:       keys,
		no_key:     no_key,
		funcAppend: getFuncAppend(no_key),
	}
}

func (lltsv *Lltsv) scanAndWrite(file *os.File, filters []string) error {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lvs := lltsv.parseLtsv(line)

		should_output := true
		for _, f := range filters {
			exp := strings.SplitN(f, " ", 3)
			switch exp[1] {
			case ">", ">=", "==", "<=", "<":
				l, _ := strconv.ParseFloat(lvs[exp[0]], 64)
				r, _ := strconv.ParseFloat(exp[2], 64)
				switch exp[1] {
				case ">":
					if !(l > r) {
						should_output = false
						break
					}
				case ">=":
					if !(l >= r) {
						should_output = false
						break
					}
				case "==":
					if !(l == r) {
						should_output = false
						break
					}
				case "<=":
					if !(l <= r) {
						should_output = false
						break
					}
				case "<":
					if !(l < r) {
						should_output = false
						break
					}
				}
			case "=~", "!~":
				l := lvs[exp[0]]
				r := exp[2]
				switch exp[1] {
				case "=~":
					if m, _ := regexp.MatchString(r, l); m == false {
						should_output = false
						break
					}
				case "!~":
					if m, _ := regexp.MatchString(r, l); m == true {
						should_output = false
						break
					}
				}
			}
		}

		if should_output {
			ltsv := lltsv.restructLtsv(lvs)
			os.Stdout.WriteString(ltsv + "\n")
		}
	}
	return scanner.Err()
}

// lvs: label and value pairs
func (lltsv *Lltsv) restructLtsv(lvs map[string]string) string {
	// specified keys or all keys
	orders := lltsv.keys
	if len(lltsv.keys) == 0 {
		orders = keysInMap(lvs)
	}
	// make slice with enough capacity so that append does not newly create object
	// cf. http://golang.org/pkg/builtin/#append
	selected := make([]string, 0, len(orders))
	for _, label := range orders {
		value := lvs[label]
		selected = lltsv.funcAppend(selected, label, value)
	}
	return strings.Join(selected, "\t")
}

func (lltsv *Lltsv) parseLtsv(line string) map[string]string {
	columns := strings.Split(line, "\t")
	lvs := make(map[string]string)
	for _, column := range columns {
		l_v := strings.SplitN(column, ":", 2)
		if len(l_v) < 2 {
			continue
		}
		label, value := l_v[0], l_v[1]
		lvs[label] = value
	}
	return lvs
}

// Return function pointer to avoid `if` evaluation occurs in each iteration
func getFuncAppend(no_key bool) tFuncAppend {
	if no_key {
		return func(selected []string, label string, value string) []string {
			return append(selected, value)
		}
	} else {
		if termutil.Isatty(os.Stdout.Fd()) {
			return func(selected []string, label string, value string) []string {
				return append(selected, ansi.Color(label, "green")+":"+ansi.Color(value, "magenta"))
			}
		} else {
			// if pipe or redirect
			return func(selected []string, label string, value string) []string {
				return append(selected, label+":"+value)
			}
		}
	}
}

func keysInMap(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
