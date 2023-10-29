package main

import (
	"archive/zip"
	"fmt"
	"github.com/sunshineplan/imgconv"
	"golang.org/x/image/webp"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"
	"time"
)

const (
	JPG  string = "JPG"
	JPEG string = "JPEG"
	PNG  string = "PNG"
	GIF  string = "GIF"
	WEBP string = "WEBP"
)

func readAndConvertFiles(dir string) ([]string, error) {
	var imgFiles []string
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		fname := file.Name()
		if len(strings.Split(fname, ".")) == 1 {
			log.Printf("error while trying to split filename: %s", fname)
			continue
		}
		extension := strings.ToUpper(strings.Split(fname, ".")[1])
		if extension == PNG || extension == WEBP {
			err = convertTransparentToJpeg(fname, extension)
			if err != nil {
				log.Printf("error while converting %s from png to jpg with error: %v", fname, err)
			} else {
				imgFiles = append(imgFiles, strings.Split(fname, ".")[0]+".jpg")
			}
		}
		if extension == JPG || extension == JPEG || extension == GIF {
			imgFiles = append(imgFiles, fname)
		}
	}
	return imgFiles, nil
}

func convertTransparentToJpeg(fname, extension string) error {
	imgFile, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer imgFile.Close()
	var imgSrc image.Image
	if extension == PNG {
		// create image from PNG file
		imgSrc, err = png.Decode(imgFile)
		if err != nil {
			return err
		}
	} else {
		imgSrc, err = webp.Decode(imgFile)
		if err != nil {
			return err
		}
	}
	// create a new Image with the same dimension of PNG image
	newImg := image.NewRGBA(imgSrc.Bounds())

	// we will use white background to replace PNG's transparent background

	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// paste PNG image OVER to newImage
	draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)

	// create new out JPEG file
	jpgImgFile, err := os.Create(strings.Split(fname, ".")[0] + ".jpg")
	if err != nil {
		return err
	}
	defer jpgImgFile.Close()
	var opt jpeg.Options
	opt.Quality = 80

	// convert newImage to JPEG encoded byte and save to jpgImgFile
	// with quality = 80
	err = jpeg.Encode(jpgImgFile, newImg, &opt)

	if err != nil {
		return err
	}

	err = os.Remove(fname)
	if err != nil {
		log.Printf("failed to delete file: %s with error: %v", fname, err)
	}
	return nil
}

func createArchive(files []string) error {
	file, err := os.Create(fmt.Sprintf("Konvertierte_Bilder_%s.zip", time.Now().Format(time.DateTime)))
	if err != nil {
		return err
	}
	defer file.Close()

	// Add files to the zip file
	wr := zip.NewWriter(file)
	defer wr.Close()
	for _, fname := range files {
		f, err := wr.Create(fname)
		if err != nil {
			log.Printf("could not add file: %s to archive with error: %v", fname, err)
			continue
		}
		imgBody, err := os.ReadFile(fname)
		if err != nil {
			log.Printf("could not add file: %s to archive with error: %v", fname, err)
			continue
		}
		_, err = f.Write(imgBody)
		if err != nil {
			log.Printf("could not add file: %s to archive with error: %v", fname, err)
		}
	}
	return nil
}

func main() {
	directory, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to read current working directory with err: %v", err)
	}
	files, err := readAndConvertFiles(directory)
	if err != nil {
		log.Fatalf("unable to read files from current working directory with err: %v", err)
	}

	var convertedFiles []string
	for _, file := range files {
		src, err := imgconv.Open(file)
		if err != nil {
			log.Printf("unable to open file: %s with err: %v", file, err)
		}

		// Resize the image to width = 1500px preserving the aspect ratio.
		dst := imgconv.Resize(src, &imgconv.ResizeOption{Width: 1500})
		f, err := os.Create(strings.Split(file, ".")[0] + ".jpg")
		if err != nil {
			log.Printf("error while creating file with err: %v", err)
		}
		// Write the resulting image as JPEG.
		err = imgconv.Write(f, dst, &imgconv.FormatOption{Format: imgconv.JPEG})

		if err != nil {
			log.Printf("failed to write image: %v", err)
		}
		f.Close()
		convertedFiles = append(convertedFiles, file)
	}
	err = createArchive(convertedFiles)
	if err != nil {
		log.Printf("error while trying to create archive from files: %s with error: %v", convertedFiles, err)
	}
}
