package action

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"./kv_helper.go"
  "./vault_helper.go"
)

type FillError struct {
	Errors []error
}

func (fillError *FillError) Error() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintln("Fill did not exit cleanly, see errors below"))
	for _, err := range fillError.Errors {
		sb.WriteString("  " + fmt.Sprintln(err.Error()))
	}
	return sb.String()
}

type escapeType string

func (ft escapeType) IsValid() error {
	switch ft {
	case None, Single, Double, YAML:
		return nil
	}
	return errors.New("Invalid escape type")
}

const (
	None   escapeType = "No"
	Single escapeType = "Single"
	Double escapeType = "Double"
	YAML   escapeType = "YAML"
)

func RunFillTemplate(vaultHelper *helpers.VaultHelper, file string) (*bytes.Buffer, error) {
	fillError := &FillError{}
	tmpl, err := template.New(path.Base(file)).Funcs(vaultFillFuncMap(vaultHelper, fillError)).ParseFiles(file)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}

	if err := tmpl.Execute(buf, nil); err != nil {
		return nil, err
	}

	// fillErrors is not mutated until after the template is executed
	if len(fillError.Errors) != 0 {
		return nil, fillError
	}

	return buf, nil
}

func vaultFillFuncMap(client *helpers.VaultHelper, fillError *FillError) template.FuncMap {
	return template.FuncMap{
		"vault": func(path string, key string) string {
			return vaultFill(path, key, YAML, client, fillError)
		},
		"vault_yamlescape": func(path string, key string) string {
			return vaultFill(path, key, YAML, client, fillError)
		},
		"vault_noescape": func(path string, key string) string {
			return vaultFill(path, key, None, client, fillError)
		},
		"vault_singleescape": func(path string, key string) string {
			return vaultFill(path, key, Single, client, fillError)
		},
		"vault_doubleescape": func(path string, key string) string {
			return vaultFill(path, key, Double, client, fillError)
		},
	}
}

func vaultFill(path string, key string, escape escapeType, client *helpers.VaultHelper, fillError *FillError) string {
	if err := escape.IsValid(); err != nil {
		fillError.Errors = append(fillError.Errors, err)
		return err.Error()
	}
	expandedPath := os.ExpandEnv(path)

	secretString, err := client.ReadKeySimple(expandedPath, key)
	if err != nil {
		fillError.Errors = append(fillError.Errors, err)
		return err.Error()
	}

	switch escape {
	case Single:
		return strings.ReplaceAll(secretString, "'", "\\'")
	case Double:
		return strings.ReplaceAll(secretString, "\"", "\\\"")
	case YAML:
		return strings.ReplaceAll(secretString, "'", "''")
	case None:
		return secretString
	}

	fillError.Errors = append(fillError.Errors, errors.New("Error: fill template hit an unknown templating error"))
	return ""
}
