package identicon

import (
	"crypto/sha1"
	"hash"
	"image"
	"image/color"
)

type IdenticonGenerator interface {
	Generate(identity string) image.Image
}

type GeneratorOptions struct {
	Salt         string
	Hash         hash.Hash
	BlockSize    int
	Padding      int
	IconSize     int
	OutputFormat string
	Inverted     bool
}

type Generator struct {
	Foreground []color.NRGBA
	Background color.NRGBA
	Options    GeneratorOptions
}

func (generator *Generator) Generate(identity string) image.Image {
	digest := generator.hashData(identity)
	fg, bg := generator.pickColors(digest)
	palette := generator.makePalette(fg, bg)
	for _, rect := range generator.makeMatrix(digest) {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			for y := rect.Min.Y; y < rect.Max.Y; y++ {
				palette.Pix[y*palette.Stride+x] = 1
			}
		}
	}
	return palette
}

func (generator *Generator) makeMatrix(digest []byte) [] image.Rectangle {
	blockSize := generator.Options.BlockSize
	width := generator.Options.IconSize / (blockSize + 1)
	cols := blockSize/2 + blockSize%2
	cells := blockSize * cols
	res := make([]image.Rectangle, 0, blockSize*blockSize)
	padding := width / 2
	for i := 0; i < cells; i++ {
		if !generator.fill(i, digest) {
			continue
		}
		column := i / blockSize
		row := i / blockSize

		pt := image.Pt(padding+(column*width), padding+(row*width))
		rect := image.Rectangle{pt, image.Pt(pt.X+width, pt.Y+width)}
		if blockSize%2 == 0 && column == cols-1 {
			rect.Max.X += width
		}
		res = append(res, rect)
		if column < cols-1 {
			rect.Min.X = padding + ((blockSize - column - 1) * width)
			rect.Max.X = rect.Min.X + width
			res = append(res, rect)
		}
	}
	return res
}

func (generator *Generator) fill(n int, digest []byte) bool {
	return digest[n/8]>>(8-(n%8)+1)&1 == 1
}

func rgb(r, g, b uint8) color.NRGBA { return color.NRGBA{r, g, b, 255} }

func NewDefaultGenerator() IdenticonGenerator {
	var defaultForeground = []color.NRGBA{
		rgb(45, 79, 255),
		rgb(254, 180, 44),
		rgb(226, 121, 234),
		rgb(30, 179, 253),
		rgb(232, 77, 65),
		rgb(49, 203, 115),
		rgb(141, 69, 170),
	}
	var background = rgb(224, 224, 224)
	return &Generator{
		Foreground: defaultForeground,
		Background: background,
		Options: GeneratorOptions{
			Salt:         "",
			Hash:         sha1.New(),
			BlockSize:    5,
			Padding:      0,
			IconSize:     200,
			OutputFormat: "webp",
			Inverted:     false,
		},
	}
}

func (generator *Generator) hashData(identity string) []byte {
	h := generator.Options.Hash
	h.Write([]byte(identity))
	h.Write([]byte(generator.Options.Salt))
	return h.Sum(nil)
}

func (generator *Generator) pickColors(digest []byte) (color.NRGBA, color.NRGBA) {
	foregroundIndex := int(digest[0]) % len(generator.Foreground)
	if generator.Options.Inverted {
		return generator.Background, generator.Foreground[foregroundIndex]
	}
	return generator.Foreground[foregroundIndex], generator.Background
}

func (generator *Generator) makePalette(foreground, background color.NRGBA) *image.Paletted {
	palette := color.Palette{background, foreground}
	size := generator.Options.IconSize
	return image.NewPaletted(image.Rect(0, 0, size, size), palette)
}

func NewGenerator(foreground []color.NRGBA, background color.NRGBA, options GeneratorOptions) IdenticonGenerator {
	if options.BlockSize == 0 {
		options.BlockSize = 5
	}
	if options.IconSize == 0 {
		options.IconSize = 200
	}
	if options.Hash == nil {
		options.Hash = sha1.New()
	}
	if options.OutputFormat == "" {
		options.OutputFormat = "png"
	}
	return &Generator{
		Foreground: foreground,
		Background: background,
		Options:    options,
	}
}
