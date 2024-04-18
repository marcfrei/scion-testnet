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

var (
	testnetCryptoPaths = []string{
		"gen/certs",
		"gen/ISD1",
		"gen/trcs",
		"gen/ASff00_0_110/certs",
		"gen/ASff00_0_110/crypto",
		"gen/ASff00_0_110/keys",
		"gen/ASff00_0_111/certs",
		"gen/ASff00_0_111/crypto",
		"gen/ASff00_0_111/keys",
		"gen/ASff00_0_112/certs",
		"gen/ASff00_0_112/crypto",
		"gen/ASff00_0_112/keys",
	}
	testnetCryptoMasterKeys = []string{
		"gen/ASff00_0_110/keys/master0.key",
		"gen/ASff00_0_110/keys/master1.key",
		"gen/ASff00_0_111/keys/master0.key",
		"gen/ASff00_0_111/keys/master1.key",
		"gen/ASff00_0_112/keys/master0.key",
		"gen/ASff00_0_112/keys/master1.key",
	}
	testnetCertDirs = []string{
		"gen/ASff00_0_110/certs",
		"gen/ASff00_0_111/certs",
		"gen/ASff00_0_112/certs",
	}
	testnetGenDir   = "gen"
	testnetTRCDir   = "gen/trcs"
	testnetTopology = "topology.topo"
)

type commandPather string

func (s commandPather) CommandPath() string {
	return string(s)
}

func main() {
	var clean bool
	flag.BoolVar(&clean, "c", false, "clean")
	flag.Parse()

	for _, p := range testnetCryptoPaths {
		_ = os.RemoveAll(p)
	}
	if clean {
		return
	}
	cmd := testcrypto.Cmd(commandPather(""))
	cmd.SetArgs([]string{"-t", testnetTopology, "-o", testnetGenDir, "--as-validity", "28d"})
	stdout, stderr := os.Stdout, os.Stderr
	null, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
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
	genMasterKeyFile := func(name string) {
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
			panic(err)
		}
		defer f.Close()
		b := make([]byte, base64.StdEncoding.EncodedLen(len(x)))
		base64.StdEncoding.Encode(b, x)
		n, err = f.Write(b)
		if err != nil {
			panic(err)
		}
		if n != len(b) {
			panic("Write failed")
		}
	}
	for _, k := range testnetCryptoMasterKeys {
		genMasterKeyFile(k)
	}
	copyDir := func(src, dst string) {
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
					copyFile := func(src, dst string) {
						s, err := os.Open(src)
						if err != nil {
							log.Fatal(err)
						}
						defer s.Close()
						d, err := os.Create(dst)
						if err != nil {
							panic(err)
						}
						defer d.Close()
						_, err = d.ReadFrom(s)
						if err != nil {
							log.Fatal(err)
						}
					}
					copyFile(filepath.Join(src, n), filepath.Join(dst, n))
				}
			}
		}
	}
	for _, dst := range testnetCertDirs {
		copyDir(testnetTRCDir, dst)
	}
}
