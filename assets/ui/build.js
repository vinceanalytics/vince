const { build } = require("esbuild");

const sharedConfig = {
    entryPoints: ["src/index.ts"],
    bundle: true,
    minify: true,
};

build({
    ...sharedConfig,
    outfile: "../js/index.js",
    keepNames: true,
});

