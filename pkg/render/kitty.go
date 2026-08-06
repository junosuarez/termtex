package render

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// RenderToTerminal takes SVG content, converts it to PNG via rsvg-convert, and displays it via Kitty Graphics Protocol or icat.
func RenderToTerminal(svgContent string, writer io.Writer) error {
	// Step 1: Convert SVG to PNG using rsvg-convert
	pngBytes, err := convertSVGToPNG(svgContent)
	if err != nil {
		// Fallback to icat via temp files if direct pipe fails
		return err
	}

	// Step 2: Display PNG using Kitty Graphics Protocol escape sequences
	if isKittySupported() {
		return PrintKittyImage(writer, pngBytes)
	}

	// Fallback to kitty +kitten icat command
	return printWithIcat(pngBytes)
}

// convertSVGToPNG calls rsvg-convert to turn SVG string into PNG bytes.
func convertSVGToPNG(svgContent string) ([]byte, error) {
	cmd := exec.Command("rsvg-convert", "-f", "png")
	cmd.Stdin = stringsReader(svgContent)

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

// PrintKittyImage outputs a PNG byte slice using raw Kitty Graphics Protocol escape sequences.
func PrintKittyImage(w io.Writer, pngBytes []byte) error {
	encoded := base64.StdEncoding.EncodeToString(pngBytes)

	// Chunk encoded data into 4096 byte chunks per Kitty Graphics Protocol spec
	chunkSize := 4096
	totalLen := len(encoded)

	for i := 0; i < totalLen; i += chunkSize {
		end := i + chunkSize
		m := 1 // more chunks follow
		if end >= totalLen {
			end = totalLen
			m = 0 // last chunk
		}

		chunk := encoded[i:end]

		if i == 0 {
			// First chunk: specify format=100 (PNG), action=T (transmit and display)
			fmt.Fprintf(w, "\x1b_Ga=T,f=100,m=%d;%s\x1b\\", m, chunk)
		} else {
			// Subsequent chunks: m=1 or m=0
			fmt.Fprintf(w, "\x1b_Gm=%d;%s\x1b\\", m, chunk)
		}
	}

	fmt.Fprintln(w)
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

func stringsReader(s string) io.Reader {
	return bytes.NewReader([]byte(s))
}
