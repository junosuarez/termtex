const { mathjax } = require("mathjax-full/js/mathjax.js");
const { TeX } = require("mathjax-full/js/input/tex.js");
const { SVG } = require("mathjax-full/js/output/svg.js");
const { liteAdaptor } = require("mathjax-full/js/adaptors/liteAdaptor.js");
const { RegisterHTMLHandler } = require("mathjax-full/js/handlers/html.js");
const { AllPackages } = require("mathjax-full/js/input/tex/AllPackages.js");

const adaptor = liteAdaptor();
RegisterHTMLHandler(adaptor);

const html = mathjax.document("", {
  InputJax: new TeX({ packages: AllPackages }),
  OutputJax: new SVG({ fontCache: "local" })
});

function renderTeXToSVG(texInput, displayMode = true, fgColor = "#cdd6f4") {
  const node = html.convert(texInput, { display: displayMode });
  let svgStr = adaptor.innerHTML(node);

  // Set fill color on root svg
  if (fgColor) {
    svgStr = svgStr.replace('<svg ', `<svg fill="${fgColor}" color="${fgColor}" `);
  }

  return svgStr;
}

if (require.main === module) {
	let input = process.argv[2];
	let display = process.argv[3] !== "false";
	let color = process.argv[4] || "#cdd6f4";

	if (!input) {
		let buf = "";
		process.stdin.on("data", chunk => buf += chunk);
		process.stdin.on("end", () => {
			console.log(renderTeXToSVG(buf.trim(), display, color));
		});
	} else {
		console.log(renderTeXToSVG(input, display, color));
	}
}

module.exports = { renderTeXToSVG };
