package modules

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ErrModuleAlreadyDevelopment is an error for modules that are already in development mode.
var ErrModuleAlreadyDevelopment = errors.New("module already in development mode")

type SubModule struct {
	Path       string
	ModuleName string
}

// ListModules lists all modules recursively for given path.
func ListSubmodules(path string) ([]*SubModule, error) {
	var (
		walker     filepath.WalkFunc
		submodules []*SubModule
	)
	walker = func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if info.Name() == "go.mod" {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				defer f.Close()

				r := bufio.NewReader(f)
				line, _, err := r.ReadLine()
				if err != nil {
					return err
				}
				moduleName := strings.TrimPrefix(string(line), "module ")
				submodules = append(submodules, &SubModule{
					Path:       filepath.Dir(path),
					ModuleName: moduleName,
				})
			}
		} else {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
		}
		return nil
	}
	if err := filepath.Walk(path, walker); err != nil {
		return nil, err
	}
	return submodules, nil
}

func lockFileName(dir string) string {
	return fmt.Sprintf("%s/.nrn-lock", dir)
}

// SetDevelopmentMode replaces all neuron and neuron-extensions modules with the 'replace' clause.
func SetDevelopmentMode(module *SubModule, neuronPath, extensionsPath string) error {
	lock, err := os.OpenFile(lockFileName(module.Path), os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil && os.IsNotExist(err) {
		lock, err = os.Create(lockFileName(module.Path))
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		return ErrModuleAlreadyDevelopment
	}
	defer lock.Close()

	// Write the timestamp in the lock.
	if _, err = lock.WriteString(time.Now().String()); err != nil {
		return err
	}

	// Rename the submodule go.mod file as the copy.
	err = os.Rename(subModuleFilename(module.Path), subModuleProductionCopy(module.Path))
	if err != nil {
		return err
	}

	// Open the production copy and read it's content.
	prod, err := os.Open(subModuleProductionCopy(module.Path))
	if err != nil {
		return err
	}
	defer prod.Close()

	r := bufio.NewReader(prod)
	buf := &bytes.Buffer{}
	var toReplace []*SubModule
	for {
		line, _, err := r.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		stringLines := strings.Split(strings.TrimSpace(string(line)), " ")
		if stringLines[0] == "require" {
			stringLines = stringLines[1:]
		}
		if stringLines[0] == "github.com/neuronlabs/neuron" {
			rel, err := filepath.Rel(module.Path, neuronPath)
			if err != nil {
				return err
			}
			toReplace = append(toReplace, &SubModule{
				Path:       rel,
				ModuleName: "github.com/neuronlabs/neuron",
			})
		} else if strings.HasPrefix(stringLines[0], "github.com/neuronlabs/neuron-extensions") {
			moduleName := stringLines[0]
			rel, err := filepath.Rel(module.Path, extensionsPath+strings.TrimPrefix(moduleName, "github.com/neuronlabs/neuron-extensions"))
			if err != nil {
				return err
			}
			toReplace = append(toReplace, &SubModule{
				Path:       rel,
				ModuleName: moduleName,
			})
		}
		if _, err = fmt.Fprintln(buf, string(line)); err != nil {
			return err
		}
	}
	if len(toReplace) == 0 {
		return nil
	}

	// Write empty line.
	if _, err = fmt.Fprintln(buf); err != nil {
		return err
	}
	if _, err = fmt.Fprintln(buf, "replace ("); err != nil {
		return err
	}
	for _, replace := range toReplace {
		if _, err = fmt.Fprintf(buf, "\t%s => %s\n", replace.ModuleName, replace.Path); err != nil {
			return err
		}
	}
	if _, err = fmt.Fprintln(buf, ")"); err != nil {
		return err
	}
	// Open the
	develop, err := os.OpenFile(subModuleFilename(module.Path), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664)
	if err != nil {
		return err
	}
	defer develop.Close()
	if _, err = buf.WriteTo(develop); err != nil {
		return err
	}
	return nil
}

// SetProductionMode goes back from development mode into versioning mode.
func SetProductionMode(module *SubModule) error {
	lock, err := os.OpenFile(lockFileName(module.Path), os.O_RDWR, 0666)
	if err != nil && os.IsNotExist(err) {
		// Noting to do in this module.
		return nil
	} else if err != nil {
		return err
	}
	lock.Close()
	// Clear the lock.
	if err = os.Remove(lockFileName(module.Path)); err != nil {
		return err
	}
	// Clear development go.modules.
	if err = os.Remove(subModuleFilename(module.Path)); err != nil {
		return err
	}
	// Rename the production module copy.
	if err = os.Rename(subModuleProductionCopy(module.Path), subModuleFilename(module.Path)); err != nil {
		return err
	}
	return nil
}

// UpdateModuleVersion updates submodule version.
func UpdateModuleVersion(module *SubModule, reference, newVersion string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(dir)

	lock, err := os.OpenFile(lockFileName(module.Path), os.O_RDWR, 0666)
	if err != nil && os.IsNotExist(err) {
		// No lock here we can update the module.
	} else if err != nil {
		return err
	} else {
		lock.Close()
		return fmt.Errorf("module in development mode")
	}

	gomod, err := os.OpenFile(subModuleFilename(module.Path), os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	defer gomod.Close()

	r := bufio.NewReader(gomod)
	w := &bytes.Buffer{}
	for {
		line, _, err := r.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		split := strings.Split(strings.TrimSpace(string(line)), " ")
		// Replace the line if it contains given reference.
		if split[0] == reference {
			line = []byte(fmt.Sprintf("\t%s %s", reference, newVersion))
		}
		if _, err = fmt.Fprintln(w, string(line)); err != nil {
			return err
		}
	}
	gomod.Close()

	gomod, err = os.OpenFile(subModuleFilename(module.Path), os.O_RDWR|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer gomod.Close()
	if _, err = w.WriteTo(gomod); err != nil {
		return err
	}

	os.Chdir(module.Path)
	cmd := exec.Command("go", "mod", "tidy")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(out) > 0 {
		fmt.Println(out)
	}

	cmd = exec.Command("go", "get", reference)
	out, err = cmd.Output()
	if err != nil {
		return err
	}
	if len(out) > 0 {
		fmt.Println(out)
	}
	return nil
}

func subModuleFilename(path string) string {
	return fmt.Sprintf("%s/go.mod", path)
}

func subModuleProductionCopy(path string) string {
	return fmt.Sprintf("%s/.prod-go.mod", path)
}
