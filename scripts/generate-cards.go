package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Create assets/cards directory if it doesn't exist
	cardsDir := "assets/cards"
	if err := os.MkdirAll(cardsDir, 0755); err != nil {
		log.Fatal("Failed to create cards directory:", err)
	}

	// Generate 84 placeholder cards (typical Dixit deck size)
	for i := 1; i <= 84; i++ {
		if err := generateCard(i, cardsDir); err != nil {
			log.Printf("Failed to generate card %d: %v", i, err)
		}
	}

	fmt.Println("Generated 84 placeholder cards in", cardsDir)
}

func generateCard(cardNumber int, outputDir string) error {
	// Create image
	width, height := 300, 450
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create gradient background
	for y := 0; y < height; y++ {
		ratio := float64(y) / float64(height)
		r := uint8(100 + ratio*100)
		g := uint8(150 + ratio*50)
		b := uint8(200 + ratio*55)

		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Add border
	borderColor := color.RGBA{80, 80, 80, 255}
	for i := 0; i < 5; i++ {
		drawRect(img, i, i, width-i, height-i, borderColor)
	}

	// Add card number
	drawText(img, fmt.Sprintf("Card %d", cardNumber), width/2, height/2)

	// Save image
	filename := filepath.Join(outputDir, fmt.Sprintf("%d.jpg", cardNumber))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: 85})
}

func drawRect(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	// Top and bottom lines
	for x := x1; x < x2; x++ {
		img.Set(x, y1, col)
		img.Set(x, y2-1, col)
	}
	// Left and right lines
	for y := y1; y < y2; y++ {
		img.Set(x1, y, col)
		img.Set(x2-1, y, col)
	}
}

func drawText(img *image.RGBA, text string, centerX, centerY int) {
	// Simple text rendering by drawing rectangles for each character
	// This is a very basic implementation - in production you'd use a proper font library
	textColor := color.RGBA{255, 255, 255, 255}

	// Draw background rectangle for text
	textWidth := len(text) * 12
	textHeight := 20
	x1 := centerX - textWidth/2
	y1 := centerY - textHeight/2
	x2 := centerX + textWidth/2
	y2 := centerY + textHeight/2

	// Fill background
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
				img.Set(x, y, color.RGBA{50, 50, 50, 200})
			}
		}
	}

	// Draw text outline (very basic)
	for i, char := range text {
		charX := x1 + i*12 + 6
		charY := centerY

		// Draw a simple representation of each character
		for dy := -8; dy <= 8; dy++ {
			for dx := -4; dx <= 4; dx++ {
				x := charX + dx
				y := charY + dy
				if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
					// Very basic character shapes
					if (char >= '0' && char <= '9') || (char >= 'A' && char <= 'Z') || char == ' ' {
						if dx*dx+dy*dy <= 16 { // Rough circle for each character
							img.Set(x, y, textColor)
						}
					}
				}
			}
		}
	}
}
