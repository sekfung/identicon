package identicon

import (
	"image/png"
	"os"
	"testing"
	"time"
)

func TestGenerator_Generate(t *testing.T) {
	generator := NewDefaultGenerator()
	img := generator.Generate(time.Now().String()+"sekfung")
	file, err := os.Create("/Users/sekfung/Documents/sekfung.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}