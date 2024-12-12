package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Must supply a configuration file and a template.")
	}

	configPath := os.Args[1]
	templatePath := os.Args[2]

	f, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Error opening config file: %v.", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
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
	defer f.Close()

	ab, err := abi.JSON(f)
	if err != nil {
		log.Fatalf("Error parsing ABI file: %v.", err)
	}

	var events []TEvent

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
				GoName:       solidityNameToGo(a.Name),
				GoType:       goType,
			})
		}

		events = append(events, TEvent{
			Name:      ev.Name,
			Arguments: args,
			ID:        ev.ID,
		})
	}

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, events)
	if err != nil {
		log.Fatal(err)
	}

	out, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("Failed to format: %v", err)
	}

	fmt.Println(string(out))
}

func solidityNameToGo(s string) string {
	// TODO(elffjs): Would be nice to use Go conventions like "ID" and "URL".
	return strings.ToUpper(s[:1]) + s[1:]
}

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
