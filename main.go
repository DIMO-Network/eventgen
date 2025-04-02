package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ABI    string   `yaml:"abi"`
	Events []string `yaml:"events"`
}

//go:embed default.tmpl
var defaultTemplate string

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Must supply an event configuration file.")
	}

	configPath := os.Args[1]

	fs := flag.NewFlagSet("", flag.ExitOnError)

	tmplPath := fs.String("t", "", "Path to a template file. If not specified then a default struct template will be used.")
	pkg := fs.String("p", "main", "Package name for the generated code. Defaults to main.")
	out := fs.String("o", "", "Output path. If empty then printed to stdout.")

	err := fs.Parse(os.Args[2:])
	if err != nil {
		// Shouldn't get here?
		log.Fatalf("Failed to parse flags: %s", err)
	}

	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Error opening config file: %v.", err)
	}

	b, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		log.Fatalf("Error reading config file contents: %v.", err)
	}

	var c Config
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		log.Fatalf("Error parsing config file: %v.", err)
	}

	abiPath := filepath.Join(filepath.Dir(configPath), c.ABI)

	f, err = os.Open(abiPath)
	if err != nil {
		log.Fatalf("Error opening ABI file: %v.", err)
	}

	ab, err := abi.JSON(f)
	f.Close()
	if err != nil {
		log.Fatalf("Error parsing ABI file: %v.", err)
	}

	inp := TOverall{
		Package: *pkg,
	}

	for _, e := range c.Events {
		h := crypto.Keccak256Hash([]byte(e))
		ev, err := ab.EventByID(h)
		if err != nil {
			log.Fatalf("Couldn't find event %q in ABI %q.", e, filepath.Base(abiPath))
		}

		var args []TArgument

		for _, a := range ev.Inputs {
			goType, err := solidityTypeToGo(a.Type.String())
			if err != nil {
				log.Fatalf("Solidity type %s in event %s not supported.", a.Type, ev.Name)
			}
			args = append(args, TArgument{
				SolidityName: a.Name,
				GoName:       abi.ToCamelCase(a.Name),
				GoType:       goType,
			})
		}

		inp.Events = append(inp.Events, TEvent{
			Name:      ev.Name,
			Arguments: args,
			ID:        ev.ID,
		})
	}

	tmpl := template.New("")
	if *tmplPath == "" {
		_, err := tmpl.Parse(defaultTemplate)
		if err != nil {
			log.Fatalf("Couldn't parse default template. This should never happen.")
		}
	} else {
		_, err := tmpl.ParseFiles(*tmplPath)
		if err != nil {
			log.Fatalf("Couldn't parse template at %q.", *tmplPath)
		}
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, inp)
	if err != nil {
		log.Fatal(err)
	}

	last, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to format: %v", err)
	}

	var writ io.WriteCloser
	if *out == "" {
		writ = os.Stdout
	} else {
		writ, err = os.Create(*out)
		if err != nil {
			log.Fatalf("Failed to open output file %q: %s", *out, err)
		}
	}

	_, err = writ.Write(last)
	writ.Close()
	if err != nil {
		log.Fatalf("Failed to write output: %s", err)
	}
}

// go-ethereum's abi package has a toGoType which is, unfortunately, private.
func solidityTypeToGo(s string) (string, error) {
	switch s {
	case "uint8":
		return "uint8", nil
	case "uint256":
		return "*big.Int", nil
	case "string":
		return "string", nil
	case "address":
		return "common.Address", nil
	case "bytes":
		return "[]byte", nil
	}
	return "", fmt.Errorf("we don't support Solidity type %s", s)
}

type TOverall struct {
	Package string
	Events  []TEvent
}

type TArgument struct {
	SolidityName string
	GoName       string
	GoType       string
}

type TEvent struct {
	Name      string
	Arguments []TArgument
	ID        common.Hash
}
