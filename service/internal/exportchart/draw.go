package exportchart

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/jamesread/orgviz/internal/orgs"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	cardFill   = color.RGBA{255, 255, 255, 255}
	cardBorder = color.RGBA{204, 204, 204, 255}
	linkColor  = color.RGBA{160, 160, 160, 255}
	textColor  = color.RGBA{32, 32, 32, 255}
	mutedColor = color.RGBA{90, 90, 90, 255}
	avatarFill = color.RGBA{230, 230, 230, 255}
)

func drawChart(laid chartLayout, style orgs.Style, avatarsDir string) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, max(1, laid.w), max(1, laid.h)))
	accent := accentColor(style.CSS)
	face := textFace()

	for _, line := range laid.links {
		stroke(img, line, linkColor)
	}
	for _, card := range laid.cards {
		drawCard(img, card, style.Features.Resolve(), accent, face, avatarsDir)
	}

	return img
}

func drawCard(img *image.RGBA, card placed, features orgs.ResolvedFeatures, accent color.RGBA, face font.Face, avatarsDir string) {
	r := image.Rect(card.x, card.y, card.x+card.w, card.y+card.h)
	fill(img, r, cardFill)
	strokeRect(img, r, cardBorder)
	fill(img, image.Rect(card.x, card.y, card.x+card.w, card.y+4), accent)
	drawCardBody(img, card, features, face, avatarsDir)
}

func drawCardBody(img *image.RGBA, card placed, features orgs.ResolvedFeatures, face font.Face, avatarsDir string) {
	x := card.x + padX
	y := card.y + padY
	if features.Avatars {
		drawAvatarThumb(img, card.person, avatarsDir, image.Rect(x, y, x+avatarSize, y+avatarSize))
		x += avatarSize + avatarGap
	}

	drawPersonText(img, card.person, features, face, x, y, card.x+card.w-padX)
}

func drawAvatarThumb(img *image.RGBA, person *orgs.Person, dir string, r image.Rectangle) {
	fill(img, r, avatarFill)
	src := loadAvatar(dir, person)
	if src == nil {
		return
	}

	draw.Draw(img, r, scaleTo(src, r.Dx(), r.Dy()), image.Point{}, draw.Over)
}

func drawPersonText(img *image.RGBA, person *orgs.Person, features orgs.ResolvedFeatures, face font.Face, x, y, right int) {
	maxChars := max(8, (right-x)/charWidth)
	y = drawLines(img, wrap(person.Name, maxChars), face, textColor, x, y, nameLine)
	if features.JobTitles {
		y = drawLines(img, wrap(person.Title, maxChars), face, mutedColor, x, y, titleLine)
	}
	if features.Team {
		drawLines(img, wrap(person.Team, maxChars), face, mutedColor, x, y, teamLine)
	}
}

func drawLines(img *image.RGBA, lines []string, face font.Face, c color.RGBA, x, y, lineH int) int {
	for _, line := range lines {
		drawString(img, face, c, x, y+lineH-4, line)
		y += lineH
	}

	return y
}

func drawString(img *image.RGBA, face font.Face, c color.RGBA, x, y int, s string) {
	if face == nil || s == "" {
		return
	}

	d := &font.Drawer{Dst: img, Src: image.NewUniform(c), Face: face, Dot: fixed.P(x, y)}
	d.DrawString(s)
}

func textFace() font.Face {
	parsed, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil
	}

	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: 13, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil
	}

	return face
}

func fill(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	r = r.Intersect(img.Bounds())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func strokeRect(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	stroke(img, link{r.Min.X, r.Min.Y, r.Max.X - 1, r.Min.Y}, c)
	stroke(img, link{r.Min.X, r.Max.Y - 1, r.Max.X - 1, r.Max.Y - 1}, c)
	stroke(img, link{r.Min.X, r.Min.Y, r.Min.X, r.Max.Y - 1}, c)
	stroke(img, link{r.Max.X - 1, r.Min.Y, r.Max.X - 1, r.Max.Y - 1}, c)
}

func stroke(img *image.RGBA, line link, c color.RGBA) {
	if line.x1 == line.x2 {
		vline(img, line.x1, line.y1, line.y2, c)
		return
	}

	hline(img, line.x1, line.x2, line.y1, c)
}

func hline(img *image.RGBA, x1, x2, y int, c color.RGBA) {
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		set(img, x, y, c)
	}
}

func vline(img *image.RGBA, x, y1, y2 int, c color.RGBA) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		set(img, x, y, c)
	}
}

func set(img *image.RGBA, x, y int, c color.RGBA) {
	if image.Pt(x, y).In(img.Bounds()) {
		img.SetRGBA(x, y, c)
	}
}

func scaleTo(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sb := src.Bounds()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sx := sb.Min.X + x*sb.Dx()/w
			sy := sb.Min.Y + y*sb.Dy()/h
			dst.Set(x, y, src.At(sx, sy))
		}
	}

	return dst
}
