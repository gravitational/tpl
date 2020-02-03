package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/gravitational/trace"
	log "github.com/sirupsen/logrus"
)

var valuesFile = flag.String("values", "", "File with values")
var outputFile = flag.String("output", "", "Output file path. Defaults to stdout if unspecified")
var templateFiles stringList

func init() {
	flag.Var(&templateFiles, "template", "Template file. Can be specified multiple times")
}

func main() {
	flag.Parse()
	if *valuesFile == "" {
		log.Error("values file is required")
		flag.Usage()
	}
	if err := run(*valuesFile, *outputFile, templateFiles...); err != nil {
		log.Fatalln(err)
	}
}

func run(valuesFile, outputFile string, templateFiles ...string) error {
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
	if outputFile == "" {
		return tpl.Execute(os.Stdout, v)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, v)
	if err != nil {
		return trace.Wrap(err)
	}
	return copyReaderWithPerms(outputFile, &buf, sharedReadWriteMask)
}

func decodeValues(valuesFile io.Reader) (v interface{}, err error) {
	_, err = toml.DecodeReader(valuesFile, &v)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return v, nil
}

// String formats this list for output
func (r stringList) String() string {
	return fmt.Sprint(([]string)(r))
}

func (r *stringList) Set(v string) error {
	if strings.TrimSpace(v) == "" {
		return trace.BadParameter("value cannot be empty")
	}
	*r = append(*r, v)
	return nil
}

// copyReaderWithPerms copies the contents from src to dst atomically.
// If dst does not exist, CopyReaderWithPerms creates it with permissions perm.
// If the copy fails, CopyReaderWithPerms aborts and dst is preserved.
// Adopted with modifications from https://go-review.googlesource.com/#/c/1591/9/src/io/ioutil/ioutil.go
func copyReaderWithPerms(dst string, src io.Reader, perm os.FileMode) error {
	tmp, err := ioutil.TempFile(filepath.Dir(dst), "")
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	defer func() {
		if err == nil {
			return
		}
		if err := os.Remove(tmp.Name()); err != nil {
			log.Errorf("Failed to remove %v: %v.", tmp.Name(), err)
		}
	}()
	_, err = io.Copy(tmp, src)
	if err != nil {
		tmp.Close()
		return trace.ConvertSystemError(err)
	}
	if err = tmp.Close(); err != nil {
		return trace.ConvertSystemError(err)
	}
	if err = os.Chmod(tmp.Name(), perm); err != nil {
		return trace.ConvertSystemError(err)
	}
	err = os.Rename(tmp.Name(), dst)
	if err != nil {
		return trace.ConvertSystemError(err)
	}
	return nil
}

type stringList []string

const sharedReadWriteMask = 0666
