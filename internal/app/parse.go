package app

import (
	"encoding/json"
	"fmt"
	"github.com/worldiety/jsonptr"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var inlineRef = regexp.MustCompile(`\$ref\{(\.|\w|/|#|-)+\}`)

type File struct {
	Filename string
	Document any
}

func (f File) Resolve(ref string) string {
	uri, _ := ExtractRef(ref)
	if strings.HasPrefix(uri, "/") {
		return uri // absolute case
	}

	// generally invalid, but works for relative local refs like ./Other.yaml or Other.yaml
	return filepath.Join(filepath.Dir(f.Filename), uri)
}

func Apply(cfg Config) ([]byte, error) {
	root, err := LoadFile(cfg.Filename)
	if err != nil {
		return nil, fmt.Errorf("cannot load %s: %w", cfg.Filename, err)
	}

	if err := MergeOAIRefs(root); err != nil {
		return nil, fmt.Errorf("cannot merge oai refs: %w", err)
	}

	return json.Marshal(root.Document)
}

func LoadFile(fname string) (File, error) {
	var obj any
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		return File{}, fmt.Errorf("cannot load '%s': %w", fname, err)
	}

	if err := yaml.Unmarshal(buf, &obj); err != nil {
		return File{}, fmt.Errorf("cannot parse yaml: %w", err)
	}

	f := File{Filename: fname, Document: obj}

	if err := InterpolateInlineRef(f); err != nil {
		return f, fmt.Errorf("cannot interpolate inlined refs: %w", err)
	}

	return f, nil
}

func IsFileRef(s string) bool {
	return !strings.HasPrefix(s, "#/")
}

func ExtractRef(s string) (filename, ptr string) {
	tokens := strings.SplitN(s, "#", 2)
	if len(tokens) == 1 {
		return tokens[0], ""
	}

	return tokens[0], tokens[1]
}

func InterpolateInlineRef(file File) error {

	err := WalkTree(file.Document, func(parent map[string]any, key, value string) error {
		var expErr error
		expanded := inlineRef.ReplaceAllStringFunc(value, func(s string) string {
			ref := s[5 : len(s)-1]
			fname, ptr := ExtractRef(ref)
			extFname := file.Resolve(fname)
			f, err := LoadFile(extFname)
			if err != nil && expErr == nil {
				expErr = fmt.Errorf("cannot load referenced file '%s': %w", extFname, err)
				return "!!!" + s
			}

			refObj, err := jsonptr.Evaluate(f.Document, ptr)

			//fmt.Printf("should resolve %s@%s => %v\n", extFname, ptr, fmt.Sprintf("%v", refObj)[:20])

			if err != nil && expErr == nil {
				expErr = fmt.Errorf("cannot resolve jsonptr in %s: %w", extFname, err)
				return "!!!" + s
			}

			return fmt.Sprintf("%v", refObj)
		})

		parent[key] = expanded
		return expErr
	})

	return err
}

func MergeOAIRefs(file File) error {
	err := WalkTree(file.Document, func(parent map[string]any, key, value string) error {
		if key != "$ref" {
			return nil
		}

		if !IsFileRef(value) {
			return nil
		}

		fname, ptr := ExtractRef(value)
		extFname := file.Resolve(fname)
		f, err := LoadFile(extFname)
		if err != nil {
			return fmt.Errorf("cannot load referenced file '%s': %w", extFname, err)
		}

		if err := MergeOAIRefs(f); err != nil {
			return fmt.Errorf("cannot merge %s: %w", extFname, err)
		}

		refObj, err := jsonptr.Evaluate(f.Document, ptr)
		if err != nil {
			return fmt.Errorf("cannot resolve jsonptr in %s: %w", extFname, err)
		}

		//fmt.Printf("should resolve %s@%s => %v\n", extFname, ptr, fmt.Sprintf("%v", refObj)[:10])

		if obj, ok := refObj.(map[string]any); ok {
			for k, v := range obj {
				parent[k] = v
			}
		} else {
			return fmt.Errorf("file %s does not contain an object (%T)", extFname, refObj)
		}

		delete(parent, key)

		return nil
	})

	return err
}

func WalkTree(root any, visitor func(parent map[string]any, key, value string) error) error {
	switch obj := root.(type) {
	case map[string]any:
		for k, v := range obj {
			if str, ok := v.(string); ok {
				if err := visitor(obj, k, str); err != nil {
					return err
				}
			} else {
				if err := WalkTree(v, visitor); err != nil {
					return err
				}
			}
		}
	case []any:
		for _, item := range obj {
			if err := WalkTree(item, visitor); err != nil {
				return err
			}
		}
	}

	return nil
}
