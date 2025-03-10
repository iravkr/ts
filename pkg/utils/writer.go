// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import "os"

func SetupTestSuite() error {
	// write files
	err := WriteFile("/tmp/tls.cacrt", TLSCacrt)
	if err != nil {
		return err
	}
	err = WriteFile("/tmp/tls.crt", TLSCrt)
	if err != nil {
		return err
	}
	err = WriteFile("/tmp/tls.key", TLSKey)
	if err != nil {
		return err
	}
	return nil

}

// WriteFile writes a file with path and string
func WriteFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(content)

	if err != nil {
		return err
	}

	return nil
}
