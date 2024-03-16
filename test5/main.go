package main

import (
    "fmt"
    "image"
    "image/draw"
    "image/jpeg"
    "image/png"
    "os"
    "path/filepath"
    "strings"

    "github.com/nfnt/resize"
)

func main() {
    // Define the original image path
    originalImagePath := "original.jpg"

    // Load the original image and detect its format
    inputFile, err := os.Open(originalImagePath)
    if err != nil {
        panic(err)
    }
    defer inputFile.Close()
    img, format, err := image.Decode(inputFile)
    if err != nil {
        panic(err)
    }

    // Extract the file name stem (without extension) for the output directory
    originalDir, originalFileName := filepath.Split(originalImagePath)
    fileStem := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
    outputDir := filepath.Join(originalDir, "images", fileStem)

    // Create the output directory, including parents if necessary
    if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
        panic(err)
    }

    // Define your x-size and y-size
    xSize := 3 // Number of horizontal slices
    ySize := 2 // Number of vertical slices

    total_images := ( xSize * ySize )

    // Define the tileSize
    tileSize := 72

    // Calculate the new dimensions
    newWidth := uint(xSize * tileSize)
    newHeight := uint(ySize * tileSize)

    // Resize the image
    resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)



    // Iterate over the tiles and save each one in the designated directory
    for y := 0; y < ySize; y++ {
        for x := 0; x < xSize; x++ {
            // Define the rectangle for the current tile
            rect := image.Rect(x*tileSize, y*tileSize, (x+1)*tileSize, (y+1)*tileSize)
            tile := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
            draw.Draw(tile, tile.Bounds(), resizedImg, rect.Min, draw.Src)

            // Construct the file name based on the position in the grid
            fileName := fmt.Sprintf("%d.%s", y*xSize+x+1, format)
            tilePath := filepath.Join(outputDir, fileName)

            // Save the tile using the original format
            outFile, err := os.Create(tilePath)
            if err != nil {
                panic(err)
            }
            switch format {
            case "jpeg":
                jpeg.Encode(outFile, tile, nil)
            case "png":
                png.Encode(outFile, tile)
            default:
                fmt.Println("Unsupported image format:", format)
            }
            outFile.Close()
        }
    }
}
