package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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

	flag.Usage = printHelp
	flag.Parse()

	if showHelp {
		printHelp()
		os.Exit(0)
	}

	// Determine TeX input expression
	var texInput string
	args := flag.Args()
	if len(args) > 0 {
		texInput = strings.Join(args, " ")
	} else {
		// Read from STDIN
		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			bytes, err := io.ReadAll(os.Stdin)
			if err == nil {
				texInput = strings.TrimSpace(string(bytes))
			}
		}
	}

	if strings.TrimSpace(texInput) == "" {
		fmt.Fprintln(os.Stderr, "Error: No LaTeX input provided.")
		fmt.Fprintln(os.Stderr, "Usage: termtex \"\\frac{1}{x^2+1}\" or echo \"\\int_0^1 x dx\" | termtex")
		os.Exit(1)
	}

	// Parse TeX into AST
	astNode, err := tex.Parse(texInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing LaTeX: %v\n", err)
		os.Exit(1)
	}

	opts := tex.RenderOptions{
		FgColor:     fgColor,
		BgColor:     bgColor,
		FontSize:    fontSize,
		Padding:     padding,
		DisplayMode: displayMode,
	}

	// Generate SVG
	svgString, err := tex.RenderSVG(astNode, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering SVG: %v\n", err)
		os.Exit(1)
	}

	// Handle direct file output
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

		// Convert to PNG for file output
		cmd := exec.Command("rsvg-convert", "-f", "png", "-o", outputFile)
		cmd.Stdin = strings.NewReader(svgString)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing PNG file via rsvg-convert: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("PNG written to %s\n", outputFile)
		return
	}

	// Handle output format selection
	switch strings.ToLower(format) {
	case "svg":
		fmt.Print(svgString)
	case "text", "ascii":
		fmt.Println(render.RenderASCII(astNode))
	case "kitty", "auto":
		err := render.RenderToTerminal(svgString, os.Stdout)
		if err != nil {
			// Fallback to text rendering if terminal graphics failed
			fmt.Fprintln(os.Stderr, "Warning: Kitty graphics protocol failed, falling back to text:")
			fmt.Println(render.RenderASCII(astNode))
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", format)
		os.Exit(1)
	}
}

func printHelp() {
	helpText := `termtex - LaTeX Math Renderer for Kitty & Terminals

USAGE:
  termtex [options] "<latex_expression>"
  echo "<latex_expression>" | termtex

EXAMPLES:
  termtex "\frac{1}{x^2+1}"
  termtex "\int_0^\infty e^{-x^2} dx = \frac{\sqrt{\pi}}{2}"
  termtex "\begin{pmatrix} a & b \\ c & d \end{pmatrix}"
  termtex -o formula.svg "\sum_{i=1}^n x_i"

OPTIONS:
  -o, --output <file>    Save rendered output to file (.svg or .png)
  -f, --format <fmt>     Output format: auto, kitty, svg, text (default: auto)
  -c, --color <color>    Foreground color (default: "#cdd6f4")
  -bg, --background <c>  Background color (default: "transparent")
  -s, --size <float>     Font size in pixels (default: 32.0)
  -p, --padding <float>  Padding around math in pixels (default: 16.0)
  -d, --display          Display mode (default: true)
  -h, --help             Show this help message
`
	fmt.Print(helpText)
}
