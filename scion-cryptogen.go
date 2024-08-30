package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/scionproto/scion/scion-pki/testcrypto"
)

type commandPather string

func (s commandPather) CommandPath() string {
	return string(s)
}

func copyFile(src, dst string) {
	s, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	_, err = d.ReadFrom(s)
	if err != nil {
		log.Fatal(err)
	}
}

func copyDir(src, dst string) {
	es, err := os.ReadDir(src)
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range es {
		n := e.Name()
		if n[0] != '.' {
			if e.IsDir() {
				panic("not yet implemented")
			} else if e.Type().IsRegular() {
				copyFile(filepath.Join(src, n), filepath.Join(dst, n))
			}
		}
	}
}

func collectCryptoPaths(cryptoPaths *[]string, dir string) {
	es, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range es {
		n := e.Name()
		if n[0] != '.' {
			if e.IsDir() {
				p := filepath.Join(dir, n)
				if n == "certs" || n == "crypto" || n == "keys" || n == "trcs" {
					*cryptoPaths = append(*cryptoPaths, p)
				} else {
					collectCryptoPaths(cryptoPaths, p)
				}
			}
		}
	}
}

func genMasterKey(name string) {
	x := make([]byte, 16)
	n, err := rand.Read(x)
	if err != nil {
		panic(err)
	}
	if n != len(x) {
		panic("rand.Read failed")
	}
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	b := make([]byte, base64.StdEncoding.EncodedLen(len(x)))
	base64.StdEncoding.Encode(b, x)
	n, err = f.Write(b)
	if err != nil {
		log.Fatal(err)
	}
	if n != len(b) {
		log.Fatal(err)
	}
}

func main() {
	var clean bool
	flag.BoolVar(&clean, "c", false, "clean")
	flag.Parse()

	genDir := flag.Arg(0)
	trcDir := filepath.Join(genDir, "trcs")
	topoFile := filepath.Join(genDir, "topology.topo")

	var cryptoPaths []string

	collectCryptoPaths(&cryptoPaths, genDir)
	for _, p := range cryptoPaths {
		_ = os.RemoveAll(p)
	}
	cryptoPaths = nil

	if clean {
		return
	}

	cmd := testcrypto.Cmd(commandPather(""))
	cmd.SetArgs([]string{"-t", topoFile, "-o", genDir, "--as-validity", "28d"})
	stdout, stderr := os.Stdout, os.Stderr
	null, err := os.Open(os.DevNull)
	if err != nil {
		log.Fatal(err)
	}
	func() {
		os.Stdout, os.Stderr = null, null
		defer func() {
			os.Stdout, os.Stderr = stdout, stderr
		}()
		err = cmd.Execute()
	}()
	if err != nil {
		log.Fatalf("testcrypto failed: %v", err)
	}

	collectCryptoPaths(&cryptoPaths, genDir)
	for _, p := range cryptoPaths {
		_, q := filepath.Split(p)
		if q == "certs" {
			copyDir(trcDir, p)
		} else if q == "keys" {
			genMasterKey(filepath.Join(p, "master0.key"))
			genMasterKey(filepath.Join(p, "master1.key"))
		}
	}
}
