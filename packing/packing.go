package packing

import (
	"archive/tar"
	"github.com/tywkeene/autobd/options"
	"io"
	"os"
	"path/filepath"
)

func UnpackDir(source io.Reader) error {
	tr := tar.NewReader(source)

	for {
		header, err := tr.Next()
		if err != nil {
			return err
		} else if err == io.EOF {
			break
		}

		filename := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(filename, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
		case tar.TypeReg:
			writer, err := os.Create(filename)
			if err != nil {
				return err
			}

			io.Copy(writer, tr)
			if err = os.Chmod(filename, os.FileMode(header.Mode)); err != nil {
				return err
			}
			writer.Close()
		}
	}
	return nil
}

//addTarFile() and PackDir() are from https://github.com/pivotal-golang/archiver
//I was originally going to bring in the whole package as a dependency but it turns out
//the extractor package doesn't entirely work the way I thought. This works, so I'm putting it
//in here to avoid the whole dependency thing until I can fix it.
//TL;DR: This isn't my code.

func addTarFile(path string, name string, tw *tar.Writer) error {
	fi, err := os.Lstat(path)
	if err != nil {
		return err
	}

	link := ""
	if fi.Mode()&os.ModeSymlink != 0 {
		if link, err = os.Readlink(path); err != nil {
			return err
		}
	}

	hdr, err := tar.FileInfoHeader(fi, link)
	if err != nil {
		return err
	}
	if fi.IsDir() && !os.IsPathSeparator(name[len(name)-1]) {
		name = name + "/"
	}
	if hdr.Typeflag == tar.TypeReg && name == "." {
		hdr.Name = filepath.ToSlash(filepath.Base(path))
	} else {
		hdr.Name = filepath.ToSlash(path)
	}
	hdr.Name = filepath.ToSlash(name)
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if hdr.Typeflag == tar.TypeReg {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func PackDir(srcPath string, dest io.Writer) error {
	absolutePath, err := filepath.Abs(srcPath)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(dest)
	defer tw.Close()

	err = filepath.Walk(absolutePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(options.Config.Root, path)
		if err != nil {
			return err
		}
		return addTarFile(path, relativePath, tw)
	})

	return err
}
