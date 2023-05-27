const { build } = require("esbuild");
const sass = require('sass');
const fs = require("fs");

const sharedConfig = {
    entryPoints: ["frontend/src/index.ts"],
    bundle: true,
    minify: true,
};

build({
    ...sharedConfig,
    outfile: "assets/js/index.js",
    keepNames: true,
});

const result = sass.compile("frontend/scss/main.scss", {
    style: "compressed",
})

fs.writeFileSync("assets/css/app.css", result.css);

// copy styles
fs.copyFileSync("node_modules/@primer/css/dist/primer.css", "assets/css/primer.css")
fs.copyFileSync("node_modules/@primer/css/dist/primer.css.map", "assets/css/primer.css.map")
fs.copyFileSync("node_modules/@primer/view-components/app/assets/styles/primer_view_components.css", "assets/css/primer_view_components.css")
fs.copyFileSync("node_modules/@primer/view-components/app/assets/styles/primer_view_components.css.map", "assets/css/primer_view_components.css.map")


//copy js
fs.copyFileSync("node_modules/@primer/view-components/app/assets/javascripts/primer_view_components.js", "assets/js/primer_view_components.js")
fs.copyFileSync("node_modules/@primer/view-components/app/assets/javascripts/primer_view_components.js.map", "assets/js/primer_view_components.js.map")