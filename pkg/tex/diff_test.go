package tex

import (
	"bytes"
	"fmt"
	"image"
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

var BenchmarkSuite = []BenchmarkFormula{
	{Name: "Quadratic", TeX: `x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`},
	{Name: "Integral", TeX: `\int_0^\infty e^{-x^2} dx = \frac{\sqrt{\pi}}{2}`},
	{Name: "Summation", TeX: `\sum_{i=1}^n x_i^2`},
	{Name: "Matrix", TeX: `\begin{pmatrix} a & b \\ c & d \end{pmatrix}`},
	{Name: "RL_Update", TeX: `V(S_t) \leftarrow V(S_t) + \alpha[R_{t+1} + \gamma V(S_{t+1}) - V(S_t)]`},
}

func TestCompareWithMathJaxOracle(t *testing.T) {
	opts := DefaultRenderOptions()

	fmt.Println("\n========================================================")
	fmt.Println("   OBJECTIVE VISUAL COMPARISON: GO NATIVE vs MATHJAX   ")
	fmt.Println("========================================================")

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

		score := CompareImageSimilarity(nativeImg, oracleImg)
		fmt.Printf("Formula: %-12s | Quality Similarity Score: %6.2f%%\n", bm.Name, score*100.0)
	}
	fmt.Println("========================================================")
}

func CompareImageSimilarity(img1, img2 image.Image) float64 {
	b1 := img1.Bounds()
	b2 := img2.Bounds()

	w1, h1 := b1.Dx(), b1.Dy()
	w2, h2 := b2.Dx(), b2.Dy()

	minW := math.Min(float64(w1), float64(w2))
	minH := math.Min(float64(h1), float64(h2))

	var totalDiff float64
	var totalPixels float64

	for y := 0; y < int(minH); y++ {
		for x := 0; x < int(minW); x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()

			diffR := math.Abs(float64(r1)-float64(r2)) / 65535.0
			diffG := math.Abs(float64(g1)-float64(g2)) / 65535.0
			diffB := math.Abs(float64(b1)-float64(b2)) / 65535.0
			diffA := math.Abs(float64(a1)-float64(a2)) / 65535.0

			pixelDiff := (diffR + diffG + diffB + diffA) / 4.0
			totalDiff += pixelDiff
			totalPixels++
		}
	}

	if totalPixels == 0 {
		return 0.0
	}

	dimRatioW := math.Min(float64(w1), float64(w2)) / math.Max(float64(w1), float64(w2))
	dimRatioH := math.Min(float64(h1), float64(h2)) / math.Max(float64(h1), float64(h2))
	dimMatchFactor := (dimRatioW + dimRatioH) / 2.0

	pixelMatchFactor := 1.0 - (totalDiff / totalPixels)

	return pixelMatchFactor * dimMatchFactor
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
