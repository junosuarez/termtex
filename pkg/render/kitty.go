package render

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
)

// RenderToTerminal renders SVG to PNG and displays it as a block image in terminal.
func RenderToTerminal(svgContent string, writer io.Writer) error {
	pngBytes, err := convertSVGToPNG(svgContent, 3.5)
	if err != nil {
		return err
	}

	if isKittySupported() {
		return PrintKittyImage(writer, pngBytes)
	}

	return printWithIcat(pngBytes)
}

// RenderInlineToTerminal renders SVG to PNG and displays it inline at current cursor position.
func RenderInlineToTerminal(svgContent string, writer io.Writer) error {
	pngBytes, err := convertSVGToPNG(svgContent, 1.8)
	if err != nil {
		return err
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(pngBytes))
	cols := 3
	rows := 1
	if err == nil && cfg.Width > 0 && cfg.Height > 0 {
		cols = int(math.Max(2, math.Round(float64(cfg.Width)/9.5)))
		rows = int(math.Max(1, math.Round(float64(cfg.Height)/20.0)))
	}

	if isKittySupported() {
		return PrintKittyInline(writer, pngBytes, cols, rows)
	}

	return printWithIcat(pngBytes)
}

// convertSVGToPNG calls rsvg-convert to turn SVG string into PNG bytes with zoom scaling.
func convertSVGToPNG(svgContent string, zoom float64) ([]byte, error) {
	if zoom <= 0 {
		zoom = 3.5
	}
	zoomStr := fmt.Sprintf("%.2f", zoom)

	cmd := exec.Command("rsvg-convert", "-f", "png", "-z", zoomStr)
	cmd.Stdin = bytes.NewReader([]byte(svgContent))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("rsvg-convert failed: %w (stderr: %s)", err, stderr.String())
	}

	return out.Bytes(), nil
}

// PrintKittyImage outputs a block PNG using Kitty Graphics Protocol escape sequences.
func PrintKittyImage(w io.Writer, pngBytes []byte) error {
	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	chunkSize := 4096
	totalLen := len(encoded)

	for i := 0; i < totalLen; i += chunkSize {
		end := i + chunkSize
		m := 1
		if end >= totalLen {
			end = totalLen
			m = 0
		}

		chunk := encoded[i:end]

		if i == 0 {
			fmt.Fprintf(w, "\x1b_Ga=T,f=100,m=%d;%s\x1b\\", m, chunk)
		} else {
			fmt.Fprintf(w, "\x1b_Gm=%d;%s\x1b\\", m, chunk)
		}
	}

	fmt.Fprintln(w)
	return nil
}

// PrintKittyInline outputs an inline PNG using C=1 and advances cursor by cols spaces.
func PrintKittyInline(w io.Writer, pngBytes []byte, cols, rows int) error {
	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	chunkSize := 4096
	totalLen := len(encoded)

	for i := 0; i < totalLen; i += chunkSize {
		end := i + chunkSize
		m := 1
		if end >= totalLen {
			end = totalLen
			m = 0
		}

		chunk := encoded[i:end]

		if i == 0 {
			fmt.Fprintf(w, "\x1b_Ga=T,f=100,c=%d,r=%d,C=1,m=%d;%s\x1b\\", cols, rows, m, chunk)
		} else {
			fmt.Fprintf(w, "\x1b_Gm=%d;%s\x1b\\", m, chunk)
		}
	}

	// Advance terminal cursor by cols spaces so subsequent text doesn't overlap image
	if cols > 0 {
		fmt.Fprint(w, strings.Repeat(" ", cols))
	}

	return nil
}

// printWithIcat executes `kitty +kitten icat` using stdin.
func printWithIcat(pngBytes []byte) error {
	cmd := exec.Command("kitty", "+kitten", "icat")
	cmd.Stdin = bytes.NewReader(pngBytes)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// isKittySupported checks if current terminal environment is Kitty or supports Kitty protocol.
func isKittySupported() bool {
	term := os.Getenv("TERM")
	kittyPid := os.Getenv("KITTY_PID")
	termProgram := os.Getenv("TERM_PROGRAM")
	return kittyPid != "" || term == "xterm-kitty" || termProgram == "ghostty" || termProgram == "WezTerm"
}
