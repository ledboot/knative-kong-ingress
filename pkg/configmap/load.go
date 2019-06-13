package configmap

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// Load reads the "Data" of a ConfigMap from a particular VolumeMount.
func Load(p string) (map[string]string, error) {
	data := make(map[string]string)
	err := filepath.Walk(p, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for info.Mode()&os.ModeSymlink != 0 {
			dirname := filepath.Dir(p)
			p, err = os.Readlink(p)
			if err != nil {
				return err
			}
			if !filepath.IsAbs(p) {
				p = path.Join(dirname, p)
			}
			info, err = os.Lstat(p)
			if err != nil {
				return err
			}
		}
		if info.IsDir() {
			return nil
		}
		b, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}
		data[info.Name()] = string(b)
		return nil
	})
	return data, err
}
