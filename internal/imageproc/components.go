package imageproc

import "image"

type ComponentPoint struct {
	X int
	Y int
}

type ConnectedComponent struct {
	Pixels   []ComponentPoint
	Area     int
	MinX     int
	MaxX     int
	MinY     int
	MaxY     int
	Branches int
}

func (c ConnectedComponent) Width() int {
	if c.Area == 0 {
		return 0
	}
	return c.MaxX - c.MinX + 1
}

func (c ConnectedComponent) Height() int {
	if c.Area == 0 {
		return 0
	}
	return c.MaxY - c.MinY + 1
}

func CollectConnectedComponents(img *image.Gray, darkMax uint8) []ConnectedComponent {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return nil
	}

	visited := make([]bool, w*h)
	components := make([]ConnectedComponent, 0, 64)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.GrayAt(x, y).Y > darkMax {
				continue
			}
			idx := (y-b.Min.Y)*w + (x - b.Min.X)
			if visited[idx] {
				continue
			}

			component := collectConnectedComponent(img, b, x, y, darkMax, visited)
			if component.Area > 0 {
				components = append(components, component)
			}
		}
	}

	return components
}

func collectConnectedComponent(img *image.Gray, b image.Rectangle, sx, sy int, darkMax uint8, visited []bool) ConnectedComponent {
	w := b.Dx()
	queue := []ComponentPoint{{X: sx, Y: sy}}
	component := ConnectedComponent{
		Pixels: make([]ComponentPoint, 0, 64),
		MinX:   sx,
		MaxX:   sx,
		MinY:   sy,
		MaxY:   sy,
	}

	for len(queue) > 0 {
		p := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		idx := (p.Y-b.Min.Y)*w + (p.X - b.Min.X)
		if visited[idx] {
			continue
		}
		visited[idx] = true

		if img.GrayAt(p.X, p.Y).Y > darkMax {
			continue
		}

		component.Pixels = append(component.Pixels, p)
		component.Area++
		if p.X < component.MinX {
			component.MinX = p.X
		}
		if p.X > component.MaxX {
			component.MaxX = p.X
		}
		if p.Y < component.MinY {
			component.MinY = p.Y
		}
		if p.Y > component.MaxY {
			component.MaxY = p.Y
		}

		neighbors := 0
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := p.X+dx, p.Y+dy
				if nx < b.Min.X || nx >= b.Max.X || ny < b.Min.Y || ny >= b.Max.Y {
					continue
				}
				if img.GrayAt(nx, ny).Y > darkMax {
					continue
				}
				neighbors++
				nidx := (ny-b.Min.Y)*w + (nx - b.Min.X)
				if !visited[nidx] {
					queue = append(queue, ComponentPoint{X: nx, Y: ny})
				}
			}
		}
		if neighbors >= 3 {
			component.Branches++
		}
	}

	return component
}
