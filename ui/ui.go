package ui

//go:generate npm run build
//go:generate rm -rfv ../assets/ui/
//go:generate mkdir -p  ../assets/ui/
//go:generate cp -rfv build/ ../assets/ui/
//go:generate rm -fv ../assets/ui/asset-manifest.json
