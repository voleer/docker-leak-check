// +build windows

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type imageType struct {
	RootFS *rootFS `json:"rootfs,omitempty"`
}

type rootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids,omitempty"`
}

type layerDBItem struct {
	ID      string
	diff    string
	cacheID string
	visited bool
}

type rawLayerType struct {
	ID      string
	visited bool
}

func folderexists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func main() {
	var folder string
	flag.StringVar(&folder, "folder", "", "Root of the Docker runtime (default \"C:\\ProgramData\\docker\")")
	flag.Parse()
	if folder == "" {
		fmt.Println("Error: folder must be supplied")
		os.Exit(-1)
	}
	if folderexists(folder) {
		imageDBFolder := filepath.Join(folder, "image", "windowsfilter", "imagedb", "content", "sha256")
		if !folderexists(imageDBFolder) {
			fmt.Printf("Error: incorrect folder structure: expected %s to exist\n", imageDBFolder)
			os.Exit(-1)
		}

		layerDBFolder := filepath.Join(folder, "image", "windowsfilter", "layerdb", "sha256")
		if !folderexists(layerDBFolder) {
			fmt.Printf("Error: incorrect folder structure: expected %s to exist\n", layerDBFolder)
			os.Exit(-1)
		}
		rawLayerFolder := filepath.Join(folder, "windowsfilter")
		if !folderexists(rawLayerFolder) {
			fmt.Printf("Error: incorrect folder structure: expected %s to exist\n", rawLayerFolder)
			os.Exit(-1)
		}
		rawLayerMap, err := createRawLayerMap(rawLayerFolder)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		layerMap, err := populateLayerDBMap(layerDBFolder)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		err = verifyImages(imageDBFolder, layerMap, rawLayerMap)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		for _, layer := range layerMap {
			if layer.visited == false {
				fmt.Println("Error: layer not referenced: ", layer.ID)
				os.Exit(-1)
			}
		}

		for _, rawLayer := range rawLayerMap {
			if rawLayer.visited == false {
				fmt.Println("Error: rawLayer not referenced: ", rawLayer.ID)
				os.Exit(-1)
			}
		}
	} else {
		fmt.Println("Error: folder does not exist")
		os.Exit(-1)
	}
}

func createRawLayerMap(rawLayerFolder string) (map[string]*rawLayerType, error) {
	files, err := ioutil.ReadDir(rawLayerFolder)
	if err != nil {
		return nil, fmt.Errorf("Error: failed to read files in %s: %v", rawLayerFolder, err)
	}
	var rawLayerMap = make(map[string]*rawLayerType)
	for _, f := range files {
		if f.IsDir() {
			rawLayer := &rawLayerType{}
			rawLayer.ID = f.Name()
			rawLayerMap[rawLayer.ID] = rawLayer
		}
	}
	return rawLayerMap, nil
}

func populateLayerDBMap(layerDBFolder string) (map[string]*layerDBItem, error) {
	// enumerate the existing layers in the LayerDB
	files, err := ioutil.ReadDir(layerDBFolder)
	if err != nil {
		return nil, fmt.Errorf("Error: failed to read files in %s: %v", layerDBFolder, err)
	}
	var layerMap = make(map[string]*layerDBItem)
	for _, f := range files {
		if f.IsDir() {
			layer := &layerDBItem{}
			layer.ID = f.Name()

			diffFile := filepath.Join(layerDBFolder, f.Name(), "diff")
			dat, err := ioutil.ReadFile(diffFile)
			if err != nil {
				return nil, fmt.Errorf("Error: failed to read file %s: %v", diffFile, err)
			}
			layer.diff = string(dat)

			cacheIDFile := filepath.Join(layerDBFolder, f.Name(), "cache-id")
			dat, err = ioutil.ReadFile(cacheIDFile)
			if err != nil {
				return nil, fmt.Errorf("Error: failed to read file %s: %v", cacheIDFile, err)
			}
			layer.cacheID = string(dat)

			layerMap[layer.diff] = layer
		}
	}
	return layerMap, nil
}

func verifyLayersOfImage(imagePath string, layerMap map[string]*layerDBItem, rawLayerMap map[string]*rawLayerType) error {
	dat, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("Error: failed to read file %s: %v", imagePath, err)
	}
	image := &imageType{}
	if err := json.Unmarshal(dat, image); err != nil {
		return fmt.Errorf("Error: failed to read JSON contents of %s: %v", imagePath, err)
	}
	for _, diff := range image.RootFS.DiffIDs {
		layer := layerMap[diff]
		if layer == nil {
			return fmt.Errorf("Error: expected layer with diff %s", diff)
		}
		if rawLayerMap[layer.cacheID] == nil {
			return fmt.Errorf("Error: expected on-disk layer %s\n", layer.cacheID)
		}
		rawLayerMap[layer.cacheID].visited = true
		layer.visited = true
	}
	return nil
}

func verifyImages(imageDBFolder string, layerMap map[string]*layerDBItem, rawLayerMap map[string]*rawLayerType) error {
	files, err := ioutil.ReadDir(imageDBFolder)
	if err != nil {
		return fmt.Errorf("Error: failed to read files in %s: %v", imageDBFolder, err)
	}
	for _, f := range files {
		if !f.IsDir() {
			imagePath := filepath.Join(imageDBFolder, f.Name())
			err := verifyLayersOfImage(imagePath, layerMap, rawLayerMap)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
