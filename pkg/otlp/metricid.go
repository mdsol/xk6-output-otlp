package otlp

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	providerFilename  = "provider_id"
	runFilename       = "run_id"
	providerIDPattern = "^[0-9a-zA-Z_]{3,16}$"
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
		err      error
		id       string
		prid     string
		rid      int
		file     *os.File
		comments []string
	)

	if _, err = os.Stat(varPrefix); os.IsNotExist(err) {
		err = os.MkdirAll(varPrefix, 0755)
		if err != nil {
			return nil, err
		}
	}

	path := filepath.Join(varPrefix, providerFilename)

	prid, _, err = readID(path, true)
	if err != nil {
		if os.IsNotExist(err) {
			prid = string(newProviderID())
			comments = []string{"# This is an auto-generated provider_id", ""}
			err = os.WriteFile(path, []byte(strings.Join(append(comments, prid), "\n")), 0666)
		}

		if err != nil {
			return nil, err
		}
	}

	path = filepath.Join(varPrefix, runFilename)

	id, comments, err = readID(path, false)
	if err != nil {
		rid = 1
	} else {
		rid, err = strconv.Atoi(string(id))
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

	if len(comments) == 0 {
		comments = append(comments, "# This is a cycling counter in range 0..255", "")
	}
	_, err = file.WriteString(strings.Join(append(comments, strconv.Itoa(rid)), "\n"))
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

func readID(path string, provider bool) (string, []string, error) {
	comments := []string{}
	file, err := os.Open(path)
	if err != nil {
		return "", comments, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id := scanner.Text()
		if id == "" || strings.HasPrefix(id, "#") {
			// Preserve leading comments
			comments = append(comments, id)
			continue
		}

		if provider {
			// No longer than 16 chars
			if len(id) > 16 {
				id = id[0:16]
			}

			if rx := regexp.MustCompile(providerIDPattern); !rx.Match([]byte(id)) {
				return "", comments, fmt.Errorf("incorrect provider_id, must match /%s/ pattern", providerIDPattern)
			}
		}

		return id, comments, nil
	}

	if err = scanner.Err(); err != nil {
		return "", comments, nil
	}

	return "", comments, fmt.Errorf("no correct provider_id found in %s file", path)
}
