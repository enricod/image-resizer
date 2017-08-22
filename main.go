package main

import (
	"log"
	"github.com/howeyc/fsnotify"
	"flag"
	"os"
	"image/jpeg"
	"github.com/nfnt/resize"
	"strings"
	"github.com/oliamb/cutter"
	"image"
)

func main() {

	// LETTURA DEI PARAMETRI APPLICAZIONE
	dir      := flag.String("dir", ".", "directory to be monitored")
	out      := flag.String("out", "../img-resized", "directory output")
	dim      := flag.Uint  ("dim", 1024, "image dimension")
	thumbDim := flag.Uint  ("thumbDim", 200, "square thumbnail dimension")
	flag.Parse()


	log.Printf("monitoring %v, OUT=%v\n", *dir, *out)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event:", ev)
				if ev.IsCreate() {
					log.Println("new file", ev.Name)
					ImageManipulate(ev.Name, *out, *dim, renamer("_s"), toNewSize());
					ImageManipulate(ev.Name, *out, *thumbDim, renamer("_t"), toSquare() );
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch(*dir)
	if err != nil {
		log.Fatal(err)
	}

	// Hang so program doesn't exit
	<-done

	/* ... do stuff ... */
	watcher.Close()
}

type renamerType func(string ) string

type imgTransform func(image2 image.Image, dim uint) (image.Image, error)

func toSquare() imgTransform {
	return func(img image.Image, dim uint) (image.Image, error) {
		croppedImg, err := cutter.Crop(img, cutter.Config{
			Width: 4,
			Height: 4,
			Mode: cutter.Centered,
			Options: cutter.Ratio,
		})
		m := resize.Resize(dim, 0, croppedImg, resize.Lanczos3)
		return m, err
	}
}

func toNewSize() imgTransform {
	return func(img image.Image, dim uint) (image.Image, error) {
		m := resize.Resize(dim, 0, img, resize.Lanczos3)
		return m, nil
	}
}

func renamer( postfix string ) renamerType {
	return func(filename string) string {
		return strings.Replace(filename, ".jpg",  postfix + ".jpg", 1)
	}
}

func NomeImmagine(absolutePath string) string {
	splitted := strings.Split(absolutePath, "/")
	return splitted[len(splitted)-1]
}

func ReadImageFromFileSystem( fileName string) (image.Image, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	img, err := jpeg.Decode(file)
	file.Close()
	return img, err
}


func ImageManipulate( fileName string, outDir string, dim uint, renamer renamerType, operation imgTransform) {
	nomeImmagine := NomeImmagine(fileName)
	img, err := ReadImageFromFileSystem(fileName)
	if err != nil {
		log.Fatal(err)
	}
	m, err := operation(img, dim)
	out, err := os.Create(outDir + "/" + renamer(nomeImmagine))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
}