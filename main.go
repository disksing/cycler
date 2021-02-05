package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Package struct {
	ImportPath  string
	Imports     []string
	Deps        []string
	TestImports []string
}

var pkgs = make(map[string]*Package)

// arg can be
// - import path "github.com/disksing/cyler"
// - directory path "./cycler"
// - wildcard path './...'
func goList(args []string) []string {
	if len(args) == 1 && pkgs[args[0]] != nil {
		return args
	}

	args = append([]string{"list", "-json"}, args...)
	cmd := exec.Command("go", args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("exec go ", args, "fail")
		panic(err)
	}
	var res []string
	dec := json.NewDecoder(bytes.NewReader(out))
	for dec.More() {
		var p Package
		err = dec.Decode(&p)
		if err != nil {
			fmt.Println("decode go list output error")
			panic(err)
		}
		pkgs[p.ImportPath] = &p
		res = append(res, p.ImportPath)
	}
	return res
}

var rootPkgs []string

func isRootSub(s string) bool {
	for _, r := range rootPkgs {
		if strings.HasPrefix(s, r) {
			return true
		}
	}
	return false
}

var deps = make(map[string][]string) // from -> []to

func main() {
	rootPkgs = goList(os.Args[1:])
	var subs []string
	for _, p := range rootPkgs {
		for _, pp := range pkgs[p].Deps {
			if isRootSub(pp) {
				subs = append(subs, pp)
			}
		}
	}
	goList(subs)

	for p, pkg := range pkgs {
		for _, dep := range pkg.Deps {
			if isRootSub(dep) {
				deps[p] = append(deps[p], dep)
			}
		}
	}

	for i := range rootPkgs {
		for j := i + 1; j < len(rootPkgs); j++ {
			checkDep2(rootPkgs[i], rootPkgs[j])
		}
	}
}

// all packages make a->b
func checkDep(a, b string) [][2]string {
	var res [][2]string
	for from, d := range deps {
		if strings.HasPrefix(from, a) && (!strings.HasPrefix(b, a) || from == a) {
			for _, to := range d {
				if strings.HasPrefix(to, b) && (!strings.HasPrefix(a, b) || to == b) {
					res = append(res, [2]string{from, to})
				}
			}
		}
	}
	return res
}

func checkDep2(a, b string) {
	ab, ba := checkDep(a, b), checkDep(b, a)
	if len(ab) > 0 && len(ba) > 0 {
		fmt.Printf("* %s, %s\n", a, b)
		for _, s := range ab {
			fmt.Printf(" - %s -> %s\n", s[0], s[1])
		}
		for _, s := range ba {
			fmt.Printf(" - %s <- %s\n", s[1], s[0])
		}
	}
}
