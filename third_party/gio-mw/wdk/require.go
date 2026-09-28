// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"bytes"
	"fmt"
	"gio-mw/token"
	"image"
	"io"
	"io/fs"
	"log"
	"os"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

func RequireIconWidget(data []byte) IconWidget {
	iconWidget, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return func(gtx layout.Context, foreground token.MatColor) layout.Dimensions {
		return iconWidget.Layout(gtx, foreground.AsNRGBA())
	}
}

func RequireImage(imageFile io.Reader) *widget.Image {
	decodedImage, format, err := image.Decode(imageFile)
	if err != nil {
		panic(fmt.Errorf("failed to decode image file: %v", err))
	}
	size := decodedImage.Bounds().Size()
	log.Printf("decoded iamge format: %s, size: %s\n", format, size)
	imageWidget := &widget.Image{
		Src: paint.NewImageOp(decodedImage),
		Fit: widget.Contain,
	}
	return imageWidget
}

func RequireImageFromFS(assetsFS fs.ReadFileFS, imagePath string) *widget.Image {
	imageBytes, err := assetsFS.ReadFile(imagePath)
	if err != nil {
		panic(fmt.Errorf("failed to open image file: %v", err))
	}
	return RequireImage(bytes.NewReader(imageBytes))
}

func RequireImageByPath(imagePath string) *widget.Image {
	imageFile, err := os.Open(imagePath)
	if err != nil {
		panic(fmt.Errorf("failed to open image file: %v", err))
	}
	defer func(imageFile *os.File) {
		err := imageFile.Close()
		if err != nil {
			panic(fmt.Errorf("failed to close image: %v", err))
		}
	}(imageFile)
	return RequireImage(imageFile)
}
