package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/gravitational/trace"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("Use: tpl <template-file> <values-file>")
	}

	templateFile := os.Args[1]
	valuesFile := os.Args[2]
	if err := run(templateFile, valuesFile); err != nil {
		log.Fatalln(err)
	}
}

func run(templateFile, valuesFile string) error {
	valsf, err := os.Open(valuesFile)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	defer valsf.Close()
	v, err := decodeValues(valsf)
	if err != nil {
		return trace.Wrap(err)
	}
	tmplf, err := os.Open(templateFile)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	defer tmplf.Close()
	tplContents, err := ioutil.ReadAll(tmplf)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	tpl, err := template.New("tpl").Parse(string(tplContents))
	if err != nil {
		return trace.Wrap(err)
	}
	return tpl.Execute(os.Stdout, v)
}

func decodeValues(valuesFile io.Reader) (v interface{}, err error) {
	_, err = toml.DecodeReader(valuesFile, &v)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return v, nil
}
