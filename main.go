package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// file path
const (
	AVATAR    = "assets/avatars/"
	CLUB      = "assets/clubs/"
	GRADE     = "assets/grades/"
	SPECIAL   = "assets/icons/"
	EXTENSION = ".png"
	FRAME     = "assets/frame/"
	FRAMEIMG  = "default"
	FONT      = "assets/font/mplus-1p-light.ttf"
	SAVE      = "build/cards/"
)

// // Data is ...
// type Data struct {
// 	members []Member `yaml:"data"`
// }

// Member is ..
type Member struct {
	id      string   `yaml:"id"`
	name    string   `yaml:"name"`
	grade   int      `yaml:"grade"`
	club    []string `yaml:"club"`
	special []string `yaml:"special"`
}

func drawFont(text string) image.Image {
	ftBinary, err := ioutil.ReadFile(FONT)
	ft, err := truetype.Parse(ftBinary)
	if err != nil {
		fmt.Print(err)
	}

	opt := truetype.Options{
		Size:              45,
		DPI:               0,
		Hinting:           0,
		GlyphCacheEntries: 0,
		SubPixelsX:        0,
		SubPixelsY:        0,
	}

	imageWidth := 576
	imageHeight := 128
	textTopMargin := 90

	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

	face := truetype.NewFace(ft, &opt)

	dr := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.Point26_6{},
	}

	dr.Dot.X = (fixed.I(imageWidth) - dr.MeasureString(text)) / 2
	dr.Dot.Y = fixed.I(textTopMargin)

	dr.DrawString(text)

	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)
	if err != nil {
		fmt.Print(err)
	}

	return img
}

func importImg(category string, param string) image.Image {
	path := category + param + EXTENSION
	file, err1 := os.Open(path)
	img, err2 := png.Decode(file)
	if err1 != nil || err2 != nil {
		fmt.Println(err1, err2)
	}
	return img
}

func generateCard(member Member) error {
	frameImage := importImg(FRAME, FRAMEIMG)
	gradeImage := importImg(GRADE, strconv.Itoa(member.grade))
	avatarImage := importImg(AVATAR, member.id)

	var clubImages []image.Image
	for _, club := range member.club {
		clubImages = append(clubImages, importImg(CLUB, club))
	}

	var specialImages []image.Image
	for _, spcial := range member.special {
		specialImages = append(specialImages, importImg(SPECIAL, spcial))
	}

	nameImage := drawFont(member.name)

	rgba := image.NewRGBA(gradeImage.Bounds())
	draw.Draw(rgba, gradeImage.Bounds(), gradeImage, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, frameImage.Bounds(), frameImage, image.Point{0, 0}, draw.Over)
	draw.Draw(rgba, avatarImage.Bounds(), avatarImage, image.Point{0, 0}, draw.Over)
	for i, clubImage := range clubImages {
		draw.Draw(rgba, gradeImage.Bounds(), clubImage, image.Point{-468, -108 * i}, draw.Over)
	}
	for j, specialImage := range specialImages {
		draw.Draw(rgba, gradeImage.Bounds(), specialImage, image.Point{0, -916 + 108*j}, draw.Over)
	}
	draw.Draw(rgba, nameImage.Bounds(), nameImage, image.Point{0, 0}, draw.Over)

	out, err1 := os.Create(SAVE + member.id + EXTENSION)
	err2 := png.Encode(out, rgba)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("%s | %s", err1, err2)
	}
	return nil
}

func toMember(data map[interface{}]interface{}) Member {
	res := Member{}
	for k, v := range data {
		switch k.(string) {
		case "id":
			res.id = v.(string)
		case "name":
			res.name = v.(string)
		case "grade":
			res.grade = v.(int)
		case "club":
			for _, u := range v.([]interface{}) {
				res.club = append(res.club, u.(string))
			}
		case "special":
			for _, u := range v.([]interface{}) {
				res.special = append(res.special, u.(string))
			}
		}
	}
	return res
}

func readOnSliceMap(fileBuffer []byte) ([]map[interface{}]interface{}, error) {
	data := make([]map[interface{}]interface{}, 48)
	err := yaml.Unmarshal(fileBuffer, &data)
	return data, err
}

func loadData(config string) []map[interface{}]interface{} {
	buf, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatal(err)
	}
	data, err := readOnSliceMap(buf)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func main() {
	data := loadData("./config.yaml")
	for _, attr := range data {
		member := toMember(attr)
		fmt.Printf("Generate %s ...\n", member.id)
		generateCard(member)
		fmt.Println("done")
	}
}
