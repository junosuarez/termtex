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
		fmt.Fprintln(os.Stderr, "Usage: termtex \"\\frac{1}{x^2+1}\" or termtex doc.md or echo \"Where $x$ is...\" | termtex")
		os.Exit(1)
	}

	opts := tex.RenderOptions{
		FgColor:     fgColor,
		BgColor:     bgColor,
		FontSize:    fontSize,
		Padding:     padding,
		DisplayMode: displayMode,
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

func containsDocumentDelimiters(s string) bool {
	return strings.Contains(s, "$") || strings.Contains(s, `\[`) || strings.Contains(s, `\(`)
}

func printHelp() {
	helpText := `termtex - LaTeX Math & Document Terminal Renderer

USAGE:
  termtex [flags] "<latex_expression>"
  termtex [flags] <file.md>
  cat document.md | termtex [flags]

EXAMPLES:
  termtex "\frac{1}{x^2+1}"
  termtex "Where $V(S_t)$ is the state value."
  termtex -o quadratic.png "x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}"

FLAGS:
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
