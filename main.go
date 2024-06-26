package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/mjibson/go-dsp/fft"

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
	if !(algoType == "-a" || algoType == "-d" || algoType == "-p" || algoType == "-w") {
		log.Fatal("wrong algo type, use one of [-a, -d, -p, -w]")
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
	case "-w":
		return goimagehash.WHash
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
	case goimagehash.WHash:
		return waveletHash
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

func getImageGrayData(img image.Image) []complex128 {
	bounds := img.Bounds()
	grayData := make([]complex128, bounds.Dx()*bounds.Dy())
	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			grayData[i] = complex(gray, 0)
			i++
		}
	}
	return grayData
}

func waveletHash(img image.Image) (*goimagehash.ImageHash, error) {
	// Apply FFT to get frequency components
	data := getImageGrayData(img)
	fftResult := fft.FFT(data)

	// Get absolute values of FFT results
	absValues := make([]float64, len(fftResult))
	for i, val := range fftResult {
		absValues[i] = math.Sqrt(real(val)*real(val) + imag(val)*imag(val))
	}

	// Calculate the median
	sortedValues := make([]float64, len(absValues))
	copy(sortedValues, absValues)
	sort.Float64s(sortedValues)
	median := sortedValues[len(sortedValues)/2]

	// Create hash: 1 if greater than median, 0 otherwise
	var hash uint64 = 0
	for i := 0; i < len(absValues); i++ {
		if absValues[i] > median {
			hash |= 1 << uint(i)
		}
	}

	return goimagehash.NewImageHash(hash, goimagehash.WHash), nil
}
