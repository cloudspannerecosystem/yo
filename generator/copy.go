package generator

import (
	"io"
	"os"
	"path/filepath"

	"go.mercari.io/yo/tplbin"
)

// CopyDefaultTemplates copies default templete files to dir.
func CopyDefaultTemplates(dir string) error {
	for _, tf := range tplbin.Assets.Files {
		if err := func() (err error) {
			file, err := os.OpenFile(filepath.Join(dir, tf.Name()), os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				return err
			}
			defer func() {
				if cerr := file.Close(); err == nil {
					err = cerr
				}
			}()

			_, err = io.Copy(file, tf)
			return
		}(); err != nil {
			return err
		}
	}
	return nil
}
