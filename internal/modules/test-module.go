package modules

import (
	"fmt"
	"os"
	"os/exec"
)

// TestModule tests given module.
func TestModule(module *SubModule) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(dir)
	os.Chdir(module.Path)

	cmd := exec.Command("go", "test", "./...")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("failed\t\t%s\t %s\n", module.Path, err)
		return err
	}
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	return nil
}
