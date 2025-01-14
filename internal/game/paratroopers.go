package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ystepanoff/paragopher/internal/audio"
	"github.com/ystepanoff/paragopher/internal/config"
	"github.com/ystepanoff/paragopher/internal/utils"
	"github.com/ystepanoff/paragopher/resources"
)

type Paratrooper struct {
	x, y        float32
	parachute   bool
	landed      bool
	walking     bool
	over, under *Paratrooper
}

// An ugly hack until vector.DrawFilledPath is available in Ebitengine
func DrawFilledSemicircle(
	screen *ebiten.Image,
	centerX, centerY, radius float32,
	startAngle, endAngle float32,
	clr color.Color,
) {
	segments := 180 // Number of triangles to approximate the semicircle
	angleStep := (endAngle - startAngle) / float32(segments)

	vertices := make([]ebiten.Vertex, (segments+1)*3)
	indices := make([]uint16, segments*3)

	for i := 0; i < segments; i++ {
		theta1 := float64((startAngle + float32(i)*angleStep) * math.Pi / 180)
		theta2 := float64((startAngle + float32(i+1)*angleStep) * math.Pi / 180)

		v0 := ebiten.Vertex{
			DstX:   centerX,
			DstY:   centerY,
			SrcX:   0,
			SrcY:   0,
			ColorR: float32(clr.(color.RGBA).R) / 255,
			ColorG: float32(clr.(color.RGBA).G) / 255,
			ColorB: float32(clr.(color.RGBA).B) / 255,
			ColorA: float32(clr.(color.RGBA).A) / 255,
		}

		v1 := ebiten.Vertex{
			DstX:   centerX + radius*float32(math.Cos(theta1)),
			DstY:   centerY + radius*float32(math.Sin(theta1)),
			SrcX:   0,
			SrcY:   0,
			ColorR: float32(clr.(color.RGBA).R) / 255,
			ColorG: float32(clr.(color.RGBA).G) / 255,
			ColorB: float32(clr.(color.RGBA).B) / 255,
			ColorA: float32(clr.(color.RGBA).A) / 255,
		}

		v2 := ebiten.Vertex{
			DstX:   centerX + radius*float32(math.Cos(theta2)),
			DstY:   centerY + radius*float32(math.Sin(theta2)),
			SrcX:   0,
			SrcY:   0,
			ColorR: float32(clr.(color.RGBA).R) / 255,
			ColorG: float32(clr.(color.RGBA).G) / 255,
			ColorB: float32(clr.(color.RGBA).B) / 255,
			ColorA: float32(clr.(color.RGBA).A) / 255,
		}

		vertices[i*3] = v0
		vertices[i*3+1] = v1
		vertices[i*3+2] = v2

		indices[i*3] = uint16(i * 3)
		indices[i*3+1] = uint16(i*3 + 1)
		indices[i*3+2] = uint16(i*3 + 2)
	}

	meshImg := ebiten.NewImage(1, 1)
	meshImg.Fill(config.ColourWhite)

	screen.DrawTriangles(vertices, indices, meshImg, nil)
}

func (g *Game) drawParatrooper(screen *ebiten.Image, p *Paratrooper) {
	if !p.landed && p.parachute {
		DrawFilledSemicircle(
			screen,
			p.x,
			p.y-config.ParachuteRadius,
			config.ParachuteRadius,
			-180.0,
			0.0,
			config.ColourTeal,
		)
		vector.StrokeLine(
			screen,
			p.x-config.ParachuteRadius+2.0,
			p.y-config.ParachuteRadius,
			p.x-config.ParatrooperWidth/2.0+1.0,
			p.y+1.0,
			1,
			config.ColourTeal,
			false,
		)
		vector.StrokeLine(
			screen,
			p.x+config.ParachuteRadius-2.0,
			p.y-config.ParachuteRadius,
			p.x+config.ParatrooperWidth/2.0-1.0,
			p.y+1.0,
			1,
			config.ColourTeal,
			false,
		)
		vector.DrawFilledRect(
			screen,
			p.x-config.ParatrooperWidth/2.0,
			p.y,
			config.ParatrooperWidth,
			config.ParatrooperHeight,
			config.ColourTeal,
			false,
		)
	} else {
		vector.DrawFilledRect(
			screen,
			p.x-config.ParatrooperWidth/2.0,
			p.y,
			config.ParatrooperWidth,
			config.ParatrooperHeight,
			config.ColourTeal,
			false,
		)
	}
}

func (g *Game) drawParatroopers(screen *ebiten.Image) {
	for _, p := range g.paratroopers {
		g.drawParatrooper(screen, p)
	}
}

func (g *Game) spawnParatrooper(x, y float32) {
	g.paratroopers = append(g.paratroopers, &Paratrooper{
		x:         x,
		y:         y,
		parachute: true,
		landed:    false,
	})
}

func (g *Game) updateParatroopers() {
	updated := make([]*Paratrooper, 0, len(g.paratroopers))
	for _, p := range g.paratroopers {
		if !p.landed {
			p.y += config.ParatrooperFallSpeed
			if p.y >= config.GroundY-config.ParatrooperHeight {
				p.y = config.GroundY - config.ParatrooperHeight
				p.landed = true
				p.walking = true
				p.parachute = false
			}
		} else {
			g.walk(p)
		}
		updated = append(updated, p)
	}
	g.paratroopers = updated
}

func (g *Game) walk(p *Paratrooper) {
	if g.showGameOverDialog {
		return
	}
	vx := float32(config.ParatrooperWalkSpeed)
	baseX := (config.ScreenWidth - config.BaseWidth) / 2
	if p.x > float32(config.ScreenWidth)/2.0 {
		vx = -vx
	}
	newX := p.x + vx
	if utils.Overlap1D(
		newX-config.ParatrooperWidth/2.0,
		config.ParatrooperWidth,
		baseX,
		config.BaseWidth,
	) {
		if p.y >= config.ScreenHeight-config.BaseHeight {
			p.walking = false
			return
		} else {
			pinkBaseX := (float32(config.ScreenWidth) - config.BaseWidth/3.0) / 2.0
			pinkBaseW := config.BaseWidth / 3
			if utils.Overlap1D(p.x-config.ParatrooperWidth/2.0, config.ParatrooperWidth, pinkBaseX, pinkBaseW) {
				g.gameOverSoundPlayer = audio.SoundPlayer(resources.GameOverSoundBytes)
				g.gameOverSoundPlayer.Play()
				g.showGameOverDialog = true
			}
		}
	}
	for _, q := range g.paratroopers {
		if (math.Abs(float64(q.x-p.x)) < 1e-6 &&
			math.Abs(float64(q.y-p.y)) < 1e-6) ||
			!q.landed ||
			q.walking {
			continue
		}
		if utils.Overlap1D(
			newX-config.ParatrooperWidth/2.0,
			config.ParatrooperWidth,
			q.x-config.ParatrooperWidth/2.0,
			config.ParatrooperWidth,
		) && math.Abs(float64(p.y-q.y)) < 1e-6 {
			if q.over == nil {
				p.x = q.x
				p.y = q.y - config.ParatrooperHeight
				q.over = p
				if p.under != nil {
					p.under.over = nil
				}
				p.under = q
			} else {
				p.walking = false
			}
			return
		}
	}
	p.x = newX
}
