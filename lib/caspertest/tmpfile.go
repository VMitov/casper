package caspertest

import "os"

func PrepareTmpFile(name string, data []byte) (*os.File, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	if _, err := f.Write(data); err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	return f, nil
}
