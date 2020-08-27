package modules

import (
	"fmt"
	"os"
	"os/exec"
)

// BuildModule tests given module.
func BuildModule(module *SubModule) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(dir)
	os.Chdir(module.Path)

	cmd := exec.Command("go", "build", "./...")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("failed\t\t%s\t %s\n", module.Path, err)
		return err
	}
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	fmt.Printf("ok\t\t%s\n", module.Path)
	return nil
}
