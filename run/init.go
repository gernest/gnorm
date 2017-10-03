package run

import (
	"fmt"
	"os"

	"gnorm.org/gnorm/database/drivers/postgres"
	"gnorm.org/gnorm/environ"
)

func Initialize(env environ.Values, root string) error {
	name := os.Args[2]
	f, err := getInittemplate(name)
	if err != nil {
		return err
	}
	stat, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(root, 0777)
			if err != nil {
				return err
			}
		}
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", root)
	}
	return f(root)
}

func getInittemplate(name string) (func(string) error, error) {
	switch name {
	case "postgres":
		return postgres.InitTemplates, nil
	default:
		return nil, fmt.Errorf("No init templates available for %s", name)
	}
}
