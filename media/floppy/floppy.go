// Copyright (c) 2014 VMware, Inc. All Rights Reserved.

package floppy

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-fs"
	"github.com/mitchellh/go-fs/fat"
)

// File stored on the floppy image
type File struct {
	Name string
	Data io.Reader
}

// Create a floppy image containing the given files
func Create(files ...*File) (string, error) {
	label := "ipmidata"

	f, err := ioutil.TempFile("", label)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := f.Truncate(1440 * 1024); err != nil {
		return "", err
	}

	device, err := fs.NewFileDisk(f)
	if err != nil {
		return "", err
	}

	config := &fat.SuperFloppyConfig{
		FATType: fat.FAT12,
		Label:   label,
		OEMName: label,
	}

	if err := fat.FormatSuperFloppy(device, config); err != nil {
		return "", err
	}

	fs, err := fat.New(device)
	if err != nil {
		return "", err
	}

	dir, err := fs.RootDir()
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if err := add(dir, file); err != nil {
			return "", err
		}
	}

	// Dell vmcli at least seems to care about the ext
	name := f.Name() + ".img"
	return name, os.Rename(f.Name(), name)
}

func add(dir fs.Directory, file *File) error {
	entry, err := dir.AddFile(file.Name)
	if err != nil {
		return err
	}

	dst, err := entry.File()
	if err != nil {
		return err
	}

	if _, err := io.Copy(dst, file.Data); err != nil {
		return err
	}

	return nil
}
