const { build } = require("esbuild");
const { dependencies } = require('./package.json')

const sharedConfig = {
    entryPoints: ["src/index.ts"],
    bundle: true,
    minify: true,
    keep_names: true,
    external: Object.keys(dependencies),
};

build({
    ...sharedConfig,
    outfile: "../js/index.js",
});

