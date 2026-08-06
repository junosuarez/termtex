package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"termtex/pkg/doc"
	"termtex/pkg/render"
	"termtex/pkg/tex"
)

func main() {
	var (
		outputFile  string
		format      string
		fgColor     string
		bgColor     string
		fontSize    float64
		padding     float64
		displayMode bool
		showHelp    bool
		showDemo    bool
	)

	flag.StringVar(&outputFile, "o", "", "Output file path (e.g., math.svg or math.png)")
	flag.StringVar(&outputFile, "output", "", "Output file path")
	flag.StringVar(&format, "f", "auto", "Output format: auto, kitty, svg, png, text")
	flag.StringVar(&format, "format", "auto", "Output format: auto, kitty, svg, png, text")
	flag.StringVar(&fgColor, "c", "#cdd6f4", "Foreground color (hex or CSS name)")
	flag.StringVar(&fgColor, "color", "#cdd6f4", "Foreground color")
	flag.StringVar(&bgColor, "bg", "transparent", "Background color")
	flag.StringVar(&bgColor, "background", "transparent", "Background color")
	flag.Float64Var(&fontSize, "s", 32.0, "Font size in pixels")
	flag.Float64Var(&fontSize, "size", 32.0, "Font size in pixels")
	flag.Float64Var(&padding, "p", 16.0, "Padding in pixels")
	flag.Float64Var(&padding, "padding", 16.0, "Padding in pixels")
	flag.BoolVar(&displayMode, "d", true, "Enable display mode (larger operators)")
	flag.BoolVar(&displayMode, "display", true, "Enable display mode")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showDemo, "demo", false, "Run feature showcase demo in terminal")

	flag.Usage = printHelp
	flag.Parse()

	if showHelp {
		printHelp()
		os.Exit(0)
	}

	opts := tex.RenderOptions{
		FgColor:     fgColor,
		BgColor:     bgColor,
		FontSize:    fontSize,
		Padding:     padding,
		DisplayMode: displayMode,
	}

	if showDemo {
		runDemo(opts)
		os.Exit(0)
	}

	var inputStr string
	args := flag.Args()
	if len(args) > 0 {
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err == nil {
				inputStr = string(content)
			} else {
				inputStr = strings.Join(args, " ")
			}
		} else {
			inputStr = strings.Join(args, " ")
		}
	} else {
		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			bytes, err := io.ReadAll(os.Stdin)
			if err == nil {
				inputStr = string(bytes)
			}
		}
	}

	inputStr = strings.TrimSpace(inputStr)
	if inputStr == "" {
		fmt.Fprintln(os.Stderr, "Error: No input provided.")
		fmt.Fprintln(os.Stderr, "Usage: termtex \"\\frac{1}{x^2+1}\" or termtex --demo or echo \"Where $x$ is...\" | termtex")
		os.Exit(1)
	}

	if containsDocumentDelimiters(inputStr) {
		err := doc.RenderDocument(os.Stdout, inputStr, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering document: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Generate SVG via MathJax (or native fallback)
	svgString, err := tex.RenderTeXToSVG(inputStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering SVG: %v\n", err)
		os.Exit(1)
	}

	if outputFile != "" {
		if strings.HasSuffix(outputFile, ".svg") {
			err := os.WriteFile(outputFile, []byte(svgString), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing SVG to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("SVG written to %s\n", outputFile)
			return
		}

		cmd := exec.Command("rsvg-convert", "-f", "png", "-z", "3.0", "-o", outputFile)
		cmd.Stdin = strings.NewReader(svgString)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing PNG file via rsvg-convert: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("PNG written to %s\n", outputFile)
		return
	}

	switch strings.ToLower(format) {
	case "svg":
		fmt.Print(svgString)
	case "text", "ascii":
		astNode, _ := tex.Parse(inputStr)
		fmt.Println(render.RenderASCII(astNode))
	case "kitty", "auto":
		err := render.RenderToTerminal(svgString, os.Stdout)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: Kitty graphics protocol failed, falling back to text:")
			astNode, _ := tex.Parse(inputStr)
			fmt.Println(render.RenderASCII(astNode))
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", format)
		os.Exit(1)
	}
}

func runDemo(opts tex.RenderOptions) {
	fmt.Println("\n==========================================================")
	fmt.Println("             TERMTEX FEATURE SHOWCASE DEMO                ")
	fmt.Println("==========================================================")

	demos := []struct {
		Title string
		TeX   string
	}{
		{
			Title: "1. Quadratic Formula (Fractions, Radicals & Powers)",
			TeX:   `x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`,
		},
		{
			Title: "2. Definite Gaussian Integral (Limits & Symbols)",
			TeX:   `\int_0^\infty e^{-x^2} dx = \frac{\sqrt{\pi}}{2}`,
		},
		{
			Title: "3. Infinite Basel Summation (Big Operators)",
			TeX:   `\sum_{n=1}^\infty \frac{1}{n^2} = \frac{\pi^2}{6}`,
		},
		{
			Title: "4. Matrix Environment (Grids & Parentheses)",
			TeX:   `\begin{pmatrix} a & b \\ c & d \end{pmatrix}`,
		},
		{
			Title: "5. Piecewise Cases (Braces & Text Mode)",
			TeX:   `\begin{cases} x^2 & \text{if } x \ge 0 \\ -x & \text{otherwise} \end{cases}`,
		},
	}

	for _, demo := range demos {
		fmt.Printf("\n--> %s\n", demo.Title)
		fmt.Printf("    TeX: %s\n\n", demo.TeX)
		svgStr, err := tex.RenderTeXToSVG(demo.TeX, opts)
		if err == nil {
			_ = render.RenderToTerminal(svgStr, os.Stdout)
		} else {
			fmt.Printf("    Error: %v\n", err)
		}
	}

	fmt.Println("\n--> 6. Mixed Document Interspersed Text & Inline Math")
	fmt.Println("    Text: \"Where $V(S_t)$ is the state value: $$V(S_t) \\leftarrow V(S_t) + \\alpha[R_{t+1} + \\gamma V(S_{t+1}) - V(S_t)]$$\"\n")

	docStr := "Where $V(S_t)$ is the state value at time $t$, updated via:\n\n$$V(S_t) \\leftarrow V(S_t) + \\alpha[R_{t+1} + \\gamma V(S_{t+1}) - V(S_t)]$$"
	_ = doc.RenderDocument(os.Stdout, docStr, opts)

	fmt.Println("\n==========================================================")
	fmt.Println("       Demo complete! Visit https://github.com/junosuarez/termtex")
	fmt.Println("==========================================================\n")
}

func containsDocumentDelimiters(s string) bool {
	return strings.Contains(s, "$") || strings.Contains(s, `\[`) || strings.Contains(s, `\(`)
}

func printHelp() {
	helpText := `termtex - LaTeX Math & Document Terminal Renderer

USAGE:
  termtex [flags] "<latex_expression>"
  termtex [flags] <file.md>
  termtex --demo
  cat document.md | termtex [flags]

EXAMPLES:
  termtex --demo
  termtex "\frac{1}{x^2+1}"
  termtex "Where $V(S_t)$ is the state value."
  termtex -o quadratic.png "x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}"

FLAGS:
  --demo                  Run interactive feature showcase demo
  -o, --output <file>     Save output to file (.svg or .png)
  -f, --format <format>   Output format: auto, kitty, svg, png, text (default "auto")
  -c, --color <hex>       Foreground text color (default "#cdd6f4")
  -bg, --background <hex> Background color (default "transparent")
  -s, --size <pixels>     Font size in pixels (default 32)
  -p, --padding <pixels>  Canvas padding in pixels (default 16)
  -d, --display           Enable display mode (default true)
  -h, --help              Show this help message
`
	fmt.Print(helpText)
}
