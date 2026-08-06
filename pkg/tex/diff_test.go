package tex

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os/exec"
	"strings"
	"testing"
)

type BenchmarkFormula struct {
	Name string
	TeX  string
}

type DiagnosticMetrics struct {
	Name                string
	GlyphVectorCoverage float64 // % of glyphs using exact MathJax TeX vector paths (0..100)
	WidthRatio          float64 // Native Width / Oracle Width (1.0 = perfect match)
	HeightRatio         float64 // Native Height / Oracle Height (1.0 = perfect match)
	BoundingBoxMatch    float64 // Aspect ratio & scale alignment score (0..100)
	PixelSSIM           float64 // Structural Similarity Index of aligned content (0..100)
	CompositeScore      float64 // Weighted overall score for hillclimbing (0..100)
}

var BenchmarkSuite = []BenchmarkFormula{
	{Name: "Quadratic", TeX: `x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`},
	{Name: "Integral", TeX: `\int_0^\infty e^{-x^2} dx = \frac{\sqrt{\pi}}{2}`},
	{Name: "Summation", TeX: `\sum_{i=1}^n x_i^2`},
	{Name: "Matrix", TeX: `\begin{pmatrix} a & b \\ c & d \end{pmatrix}`},
	{Name: "RL_Update", TeX: `V(S_t) \leftarrow V(S_t) + \alpha[R_{t+1} + \gamma V(S_{t+1}) - V(S_t)]`},
}

func TestCompareWithMathJaxOracle(t *testing.T) {
	opts := DefaultRenderOptions()
	opts.Padding = 4.0

	fmt.Println("\n==========================================================================================")
	fmt.Println("             STRUCTURED DIAGNOSTIC BENCHMARK: GO NATIVE vs MATHJAX ORACLE                 ")
	fmt.Println("==========================================================================================")
	fmt.Printf("%-12s | %-12s | %-12s | %-12s | %-12s | %-12s\n",
		"Formula", "Glyph Cover", "BBox Match", "Width Ratio", "Pixel SSIM", "COMPOSITE")
	fmt.Println("------------------------------------------------------------------------------------------")

	var totalComposite float64
	count := 0

	for _, bm := range BenchmarkSuite {
		oracleSVG, err := RenderTeXWithMathJax(bm.TeX, opts)
		if err != nil {
			t.Logf("[%s] Skipped MathJax oracle: %v", bm.Name, err)
			continue
		}

		astNode, err := Parse(bm.TeX)
		if err != nil {
			t.Fatalf("[%s] Native Parse error: %v", bm.Name, err)
		}
		nativeSVG, err := RenderSVG(astNode, opts)
		if err != nil {
			t.Fatalf("[%s] Native RenderSVG error: %v", bm.Name, err)
		}

		oraclePNG, err := convertSVGToPNGBytes(oracleSVG)
		if err != nil {
			t.Fatalf("[%s] Oracle SVG conversion failed: %v", bm.Name, err)
		}
		nativePNG, err := convertSVGToPNGBytes(nativeSVG)
		if err != nil {
			t.Fatalf("[%s] Native SVG conversion failed: %v", bm.Name, err)
		}

		oracleImg, _, err1 := image.Decode(bytes.NewReader(oraclePNG))
		nativeImg, _, err2 := image.Decode(bytes.NewReader(nativePNG))

		if err1 != nil || err2 != nil {
			t.Fatalf("[%s] Image decode failed", bm.Name)
		}

		metrics := ComputeDiagnosticMetrics(bm.Name, bm.TeX, nativeSVG, oracleSVG, nativeImg, oracleImg)

		fmt.Printf("%-12s | %11.1f%% | %11.1f%% | %12.2f | %11.1f%% | %11.1f%%\n",
			metrics.Name,
			metrics.GlyphVectorCoverage,
			metrics.BoundingBoxMatch,
			metrics.WidthRatio,
			metrics.PixelSSIM,
			metrics.CompositeScore,
		)

		totalComposite += metrics.CompositeScore
		count++
	}

	if count > 0 {
		avgScore := totalComposite / float64(count)
		fmt.Println("------------------------------------------------------------------------------------------")
		fmt.Printf("OVERALL HILLCLIMBING SUITE COMPOSITE SCORE: %6.1f%%\n", avgScore)
	}
	fmt.Println("==========================================================================================")
}

func ComputeDiagnosticMetrics(name, texStr, nativeSVG, oracleSVG string, nativeImg, oracleImg image.Image) DiagnosticMetrics {
	glyphCoverage := calculateGlyphCoverage(texStr)

	cropNative, nativeBox := cropToContent(nativeImg)
	cropOracle, oracleBox := cropToContent(oracleImg)

	nw, nh := nativeBox.Dx(), nativeBox.Dy()
	ow, oh := oracleBox.Dx(), oracleBox.Dy()

	widthRatio := 0.0
	heightRatio := 0.0
	bboxMatch := 0.0

	if ow > 0 && oh > 0 {
		widthRatio = float64(nw) / float64(ow)
		heightRatio = float64(nh) / float64(oh)

		aspectNative := float64(nw) / float64(nh)
		aspectOracle := float64(ow) / float64(oh)

		aspectMatch := math.Min(aspectNative, aspectOracle) / math.Max(aspectNative, aspectOracle)
		sizeMatch := math.Min(float64(nw*nh), float64(ow*oh)) / math.Max(float64(nw*nh), float64(ow*oh))

		bboxMatch = (aspectMatch*0.6 + sizeMatch*0.4) * 100.0
	}

	ssim := computeNormalizedSSIM(cropNative, cropOracle) * 100.0

	composite := (bboxMatch * 0.35) + (ssim * 0.40) + (glyphCoverage * 0.25)

	return DiagnosticMetrics{
		Name:                name,
		GlyphVectorCoverage: glyphCoverage,
		WidthRatio:          widthRatio,
		HeightRatio:         heightRatio,
		BoundingBoxMatch:    bboxMatch,
		PixelSSIM:           ssim,
		CompositeScore:      composite,
	}
}

func calculateGlyphCoverage(texStr string) float64 {
	symbols := extractSymbols(texStr)
	if len(symbols) == 0 {
		return 100.0
	}
	matched := 0
	for _, sym := range symbols {
		if _, ok := GetGlyphInfo(sym); ok {
			matched++
		}
	}
	return (float64(matched) / float64(len(symbols))) * 100.0
}

func extractSymbols(texStr string) []string {
	var syms []string
	runes := []rune(texStr)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch == '\\' {
			start := i + 1
			for start < len(runes) && ((runes[start] >= 'a' && runes[start] <= 'z') || (runes[start] >= 'A' && runes[start] <= 'Z')) {
				start++
			}
			if start > i+1 {
				macro := string(runes[i:start])
				if macro != "\\begin" && macro != "\\end" && macro != "\\frac" && macro != "\\text" {
					syms = append(syms, macro)
				}
				i = start - 1
			}
		} else if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || strings.ContainsRune("+-=/*()[]{}", ch) {
			syms = append(syms, string(ch))
		}
	}
	return syms
}

func cropToContent(img image.Image) (image.Image, image.Rectangle) {
	bounds := img.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if c.A > 20 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if minX > maxX || minY > maxY {
		return img, image.Rect(0, 0, 1, 1)
	}

	rect := image.Rect(minX, minY, maxX+1, maxY+1)
	sub, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if ok {
		return sub.SubImage(rect), rect
	}

	return img, rect
}

func computeNormalizedSSIM(img1, img2 image.Image) float64 {
	gridW, gridH := 100, 40
	norm1 := sampleGrid(img1, gridW, gridH)
	norm2 := sampleGrid(img2, gridW, gridH)

	var diffSum float64
	for i := 0; i < gridW*gridH; i++ {
		diff := math.Abs(norm1[i] - norm2[i])
		diffSum += diff
	}

	return 1.0 - (diffSum / float64(gridW*gridH))
}

func sampleGrid(img image.Image, targetW, targetH int) []float64 {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	grid := make([]float64, targetW*targetH)

	if srcW == 0 || srcH == 0 {
		return grid
	}

	for ty := 0; ty < targetH; ty++ {
		sy := bounds.Min.Y + int(float64(ty)*float64(srcH)/float64(targetH))
		for tx := 0; tx < targetW; tx++ {
			sx := bounds.Min.X + int(float64(tx)*float64(srcW)/float64(targetW))
			c := color.NRGBAModel.Convert(img.At(sx, sy)).(color.NRGBA)
			grid[ty*targetW+tx] = float64(c.A) / 255.0
		}
	}
	return grid
}

func convertSVGToPNGBytes(svgContent string) ([]byte, error) {
	cmd := exec.Command("rsvg-convert", "-f", "png")
	cmd.Stdin = strings.NewReader(svgContent)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("rsvg-convert failed: %w (%s)", err, stderr.String())
	}

	return out.Bytes(), nil
}
