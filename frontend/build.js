import { build } from "esbuild";
import pkg from 'sass';
const { compile } = pkg;
import { writeFileSync, copyFileSync } from "fs";

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

const result = compile("frontend/scss/main.scss", {
    style: "compressed",
})

writeFileSync("assets/css/app.css", result.css);

// copy styles
copyFileSync("node_modules/@primer/css/dist/primer.css", "assets/css/primer.css")
copyFileSync("node_modules/@primer/css/dist/primer.css.map", "assets/css/primer.css.map")
copyFileSync("node_modules/@primer/view-components/app/assets/styles/primer_view_components.css", "assets/css/primer_view_components.css")
copyFileSync("node_modules/@primer/view-components/app/assets/styles/primer_view_components.css.map", "assets/css/primer_view_components.css.map")


//copy js
copyFileSync("node_modules/@primer/view-components/app/assets/javascripts/primer_view_components.js", "assets/js/primer_view_components.js")
copyFileSync("node_modules/@primer/view-components/app/assets/javascripts/primer_view_components.js.map", "assets/js/primer_view_components.js.map")