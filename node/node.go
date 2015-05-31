package node

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tywkeene/autobd/api"
	"github.com/tywkeene/autobd/options"
	"github.com/tywkeene/autobd/packing"
	"github.com/tywkeene/autobd/version"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

func Get(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("User-Agent", "Autobd-node/"+version.Server())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		buffer, _ := DeflateResponse(resp)
		var err string
		json.Unmarshal(buffer, &err)
		return nil, errors.New(err)
	}
	return resp, nil
}

func DeflateResponse(resp *http.Response) ([]byte, error) {
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buffer []byte
	buffer, _ = ioutil.ReadAll(reader)
	return buffer, nil
}

func WriteFile(filename string, source io.Reader) error {
	writer, err := os.Create(filename)
	if err != nil {
		log.Println(err)
		return err
	}
	defer writer.Close()

	gr, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer gr.Close()

	io.Copy(writer, gr)
	return nil
}

func RequestVersion(seed string) (*api.VersionInfo, error) {
	log.Println("Requesting version from", seed)
	url := seed + "/version"
	resp, err := Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buffer, err := DeflateResponse(resp)
	if err != nil {
		return nil, err
	}

	var ver *api.VersionInfo
	if err := json.Unmarshal(buffer, &ver); err != nil {
		return nil, err
	}
	return ver, nil
}

func RequestManifest(seed string, dir string) (map[string]*api.Manifest, error) {
	log.Printf("Requesting manifest for directory %s from %s", dir, seed)
	url := seed + "/" + version.API() + "/manifest?dir=" + dir
	resp, err := Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buffer, err := DeflateResponse(resp)
	if err != nil {
		return nil, err
	}

	manifest := make(map[string]*api.Manifest)
	if err := json.Unmarshal(buffer, &manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func RequestSync(seed string, file string) error {
	log.Printf("Requesting sync of file '%s' from %s", file, seed)
	url := seed + "/" + version.API() + "/sync?grab=" + file
	resp, err := Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") == "application/x-tar" {
		err := packing.UnpackDir(resp.Body)
		if err != nil && err == io.EOF {
			return nil
		} else {
			return err
		}
	}

	//make sure we create the directory tree if it's needed
	if tree := path.Dir(file); tree != "" {
		err := os.MkdirAll(tree, 0777)
		if err != nil {
			return err
		}
	}
	err = WriteFile(file, resp.Body)
	return err
}

func validateServerVersion(remote *api.VersionInfo) error {
	log.Println("Checking seed server's version...")
	if version.Server() != remote.Ver {
		return fmt.Errorf("Mismatched version with server. Server: %s Local: %s",
			remote.Ver, version.Server())
	}
	if version.API() != remote.Api {
		return fmt.Errorf("Mismatched API version with server. Server: %s Local: %s",
			remote.Api, version.API())
	}
	log.Println("Version OK")
	return nil
}

func UpdateLoop(config options.NodeConf) error {
	log.Printf("Running as a node... [update_interval = %s seed_server = %s]\n",
		config.UpdateInterval, config.Seed)

	remoteVer, err := RequestVersion(config.Seed)
	if err != nil {
		return err
	}
	if err := validateServerVersion(remoteVer); err != nil {
		return err
	}

	updateInterval, err := time.ParseDuration(config.UpdateInterval)
	if err != nil {
		return err
	}

	for {
		log.Printf("Updating with %s...\n", config.Seed)
		time.Sleep(updateInterval)
	}
	return nil
}
