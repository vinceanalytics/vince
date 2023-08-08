package main

import "github.com/vinceanalytics/vince/tools"

func main() {
	tools.ExecPlain("npm", "run", "build")
	tools.Remove("../assets/ui/")
	tools.CopyDir("build/", "../assets/ui/")
	tools.Remove("../assets/ui/asset-manifest.json")
}
