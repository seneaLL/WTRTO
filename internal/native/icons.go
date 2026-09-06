package native

import "github.com/seneaLL/WTRTO/internal/native/icons"

var svgIcons = map[string][][]icons.Point{
	"save":       icons.IconSave,
	"trash":      icons.IconTrash,
	"share":      icons.IconShare,
	"align-h":    icons.IconAlignH,
	"align-v":    icons.IconAlignV,
	"export":     icons.IconExport,
	"import":     icons.IconImport,
	"plus":       icons.IconPlus,
	"minus":      icons.IconMinus,
	"clipboard":  icons.IconClipboard,
	"caret-down": icons.IconCaretDown,
	"download":   icons.IconDownload,
}

func flipY(paths [][]icons.Point) [][]icons.Point {
	out := make([][]icons.Point, len(paths))
	for i, sp := range paths {
		np := make([]icons.Point, len(sp))
		for j, p := range sp {
			np[j] = icons.Point{X: p.X, Y: -p.Y}
		}
		out[i] = np
	}

	return out
}

func drawSVGIcon(c *Canvas, r Rect, points [][]icons.Point, col Color) {
	cx, cy := float64(r.X)+float64(r.W)/2, float64(r.Y)+float64(r.H)/2
	s := float64(r.W)
	if r.H < r.W {
		s = float64(r.H)
	}
	s *= 0.5

	scaled := make([][]Point, len(points))
	for i, sp := range points {
		out := make([]Point, len(sp))
		for j, p := range sp {
			out[j] = Point{X: cx + p.X*s, Y: cy + p.Y*s}
		}
		scaled[i] = out
	}
	c.FillPath(scaled, col)
}
