package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/corona10/goimagehash"
)

func main() {
	log.SetOutput(os.Stderr)
	args := os.Args
	argLength := len(args)
	if argLength < 3 {
		log.Fatal("too few arguments")
	}
	algoType := args[1]
	if !(algoType == "-a" || algoType == "-d" || algoType == "-p") {
		log.Fatal("wrong algo type, use one of [-a, -d, -p]")
	}
	if argLength == 4 {
		imagePath1, imagePath2 := args[2], args[3]
		hash1, hash2, distance, err := distanceModeAction(imagePath1, imagePath2, ImageHashKindStringToKind(algoType))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d %d %d\n", hash1, hash2, distance)
		os.Exit(0)
	}
	path := args[2]
	hash, err := singleCalculationModeAction(path, ImageHashKindStringToKind(algoType))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d\n", hash)
	os.Exit(0)
}

func ImageHashKindStringToKind(algorithmType string) goimagehash.Kind {
	switch algorithmType {
	case "-d":
		return goimagehash.DHash
	case "-p":
		return goimagehash.PHash
	case "-a":
		return goimagehash.AHash
	default:
		return goimagehash.Unknown
	}
}

func ImageHashKindToFunc(algorithmType goimagehash.Kind) func(img image.Image) (*goimagehash.ImageHash, error) {
	switch algorithmType {
	case goimagehash.DHash:
		return goimagehash.DifferenceHash
	case goimagehash.PHash:
		return goimagehash.PerceptionHash
	case goimagehash.AHash:
		return goimagehash.AverageHash
	default:
		return nil
	}
}

func singleCalculationModeAction(path string, algorithmType goimagehash.Kind) (uint64, error) {
	imageData, _, err := openFile(path)
	if err != nil {
		return 0, err
	}
	result, err := ImageHashKindToFunc(algorithmType)(*imageData)
	return result.GetHash(), err
}

func distanceModeAction(path1, path2 string, algorithmType goimagehash.Kind) (uint64, uint64, int, error) {
	if algorithmType == goimagehash.Unknown {
		return 0, 0, 0, errors.New("bad algorithm Kind")
	}
	hashFunc := ImageHashKindToFunc(algorithmType)
	results := make([]goimagehash.ImageHash, 0, 2)
	for _, path := range []string{path1, path2} {
		imageData, _, err := openFile(path)
		if err != nil {
			return 0, 0, 0, err
		}
		result, err := hashFunc(*imageData)
		if err != nil {
			return 0, 0, 0, err
		}
		results = append(results, *result)
	}
	distance, err := results[0].Distance(&results[1])
	if err != nil {
		return 0, 0, 0, err
	}
	return results[0].GetHash(), results[1].GetHash(), distance, nil
}

func openFile(path string) (*image.Image, string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return imageFromURL(path)
	}
	return imageFromLocal(path)
}

func imageFromLocal(path string) (*image.Image, string, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, "", err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()
	imgData, format, err := image.Decode(file)
	if err != nil {
		log.Printf("file: %s format: %s error: %s\n", path, format, err)
		return nil, "", err
	}
	return &imgData, format, nil
}

func imageFromURL(path string) (*image.Image, string, error) {
	response, err := http.DefaultClient.Get(path)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	imgData, format, err := image.Decode(response.Body)
	if err != nil {
		log.Printf("file: %s format: %s error: %s\n", path, format, err)
		return nil, "", err
	}
	return &imgData, format, nil
}
