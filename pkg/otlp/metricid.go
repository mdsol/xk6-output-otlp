package otlp

import (
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

const (
	providerFilename = "provider_id"
	runFilename      = "run_id"
)

var varPrefix string

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	varPrefix = filepath.Join(usr.HomeDir, ".xk6-output-otlp/")
}

type idAttrs struct {
	providerID string
	runID      byte
}

func newIdentities() (*idAttrs, error) {
	var (
		err  error
		data []byte
		prid string
		rid  int
		file *os.File
	)

	if _, err = os.Stat(varPrefix); os.IsNotExist(err) {
		err = os.MkdirAll(varPrefix, 0755)
		if err != nil {
			return nil, err
		}
	}

	path := filepath.Join(varPrefix, providerFilename)

	data, err = os.ReadFile(path)

	if err != nil {
		if os.IsNotExist(err) {
			data = newProviderID()
			err = os.WriteFile(path, data, 0666)
		}

		if err != nil {
			return nil, err
		}
	}

	prid = string(data)

	path = filepath.Join(varPrefix, runFilename)

	data, err = os.ReadFile(path)
	if err != nil {
		rid = 1
	} else {
		rid, err = strconv.Atoi(string(data))
		if err != nil {
			rid = 1
		} else {
			rid = (rid + 1) % 256
		}
	}

	file, err = os.Create(path)
	if err != nil {
		return nil, err
	}

	_, err = file.WriteString(strconv.Itoa(rid))
	if err != nil {
		return nil, err
	}

	retval := &idAttrs{
		providerID: string(prid),
		runID:      byte(rid),
	}

	return retval, nil
}

func newProviderID() []byte {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	retval := make([]byte, 8)
	for i := range retval {
		retval[i] = byte(letters[rand.Intn(len(letters))])
	}

	return retval
}
