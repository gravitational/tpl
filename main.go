package main

import (
	"io"
	"log"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/gravitational/trace"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalln("Use: tpl <values-file> <template-file> [<template-file>...]")
	}

	valuesFile := os.Args[1]
	if err := run(valuesFile, os.Args[2:]...); err != nil {
		log.Fatalln(err)
	}
}

func run(valuesFile string, templateFiles ...string) error {
	valsf, err := os.Open(valuesFile)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	defer valsf.Close()
	v, err := decodeValues(valsf)
	if err != nil {
		return trace.Wrap(err)
	}
	tpl, err := template.ParseFiles(templateFiles...)
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
