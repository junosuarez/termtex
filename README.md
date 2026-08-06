# termtex

`termtex` is a lightweight, pure Go CLI tool that renders LaTeX math expressions directly into the terminal window using vector SVG typesetting and the **Kitty Graphics Protocol**.

It focuses on standard mathematical equations, fractions, square roots, integrals, summations, matrices, and piecewise cases without requiring MathJax or heavy LaTeX installation.

![Quadratic Formula](assets/quadratic.png)

---

## ✨ Features

- **Pure Go Parser & Typesetter**: Built-in recursive-descent TeX parser and box layout engine.
- **Kitty Graphics Protocol**: Crisp, pixel-perfect math rendering directly in Kitty, Ghostty, WezTerm, and compatible terminals.
- **Vector SVG & PNG Export**: Output cleanly to `.svg` or `.png` files.
- **Unicode Terminal Fallback**: Terminal text fallback mode (`-f text`) for standard environments.
- **Standard Math Subset**:
  - Fractions: `\frac{a}{b}`
  - Square & N-th Roots: `\sqrt{x}`, `\sqrt[n]{x}`
  - Subscripts & Superscripts: `x_i`, `e^{-x^2}`, `\sum_{i=1}^n`
  - Big Operators: `\sum`, `\prod`, `\int`, `\iint`, `\iiint`, `\oint`, `\lim`, `\max`, `\min`
  - Delimiters & Braces: `\left( ... \right)`, `\left[ ... \right]`, `\left\{ ... \right\}`, `\left\langle ... \right\rangle`
  - Environments: `\begin{pmatrix}`, `\begin{bmatrix}`, `\begin{cases}`
  - Greek Letters & Accents: `\alpha`, `\beta`, `\Delta`, `\hat{x}`, `\vec{v}`, `\bar{z}`

---

## 📸 Output Examples

### Definite Integral & Square Root
`\int_0^\infty e^{-x^2} dx = \frac{\sqrt{\pi}}{2}`

![Definite Integral](assets/integral.png)

### Piecewise Functions
`\begin{cases} x & \text{if } x > 0 \\ -x & \text{otherwise} \end{cases}`

![Piecewise Cases](assets/cases.png)

### Matrix Layout
`\begin{pmatrix} a & b \\ c & d \end{pmatrix}`

![Matrix](assets/matrix.png)

---

## 🚀 Installation

### Prerequisites
- **Go 1.20+**
- **`rsvg-convert`** (for SVG-to-PNG terminal rasterization):
  - macOS: `brew install librsvg`
  - Linux (Ubuntu/Debian): `sudo apt install librsvg2-bin`

### Install via `go install`
```bash
go install github.com/junosuarez/termtex@latest
```

### Build from Source
```bash
git clone https://github.com/junosuarez/termtex.git
cd termtex
go build -o termtex main.go
```

---

## 🛠️ Usage

```bash
# Render TeX equation directly to terminal
termtex "\frac{1}{x^2+1}"

# Pipe LaTeX from STDIN
echo "\sum_{i=1}^n x_i" | termtex

# Save SVG or PNG output to file
termtex -o formula.svg "\frac{a}{b}"
termtex -o formula.png "\sqrt[3]{x^2+y^2}"

# Render Unicode text fallback
termtex -f text "\int_0^\infty e^{-x^2} dx"
```

### CLI Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-o, --output <file>` | Save rendered output to `.svg` or `.png` file | `""` |
| `-f, --format <fmt>` | Output format: `auto`, `kitty`, `svg`, `png`, `text` | `"auto"` |
| `-c, --color <hex>` | Math foreground color | `"#cdd6f4"` |
| `-bg, --background <c>` | Canvas background color | `"transparent"` |
| `-s, --size <float>` | Base font size in pixels | `32.0` |
| `-p, --padding <float>` | Canvas padding around equation | `16.0` |
| `-d, --display` | Enable display mode layout | `true` |

---

## 📄 License

This project is dedicated to the public domain under the [CC0 1.0 Universal Public Domain Dedication](LICENSE).
