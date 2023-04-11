package templates

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/belak/octicon"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/flash"
	"github.com/gernest/vince/internal/plans"
	"github.com/gernest/vince/models"
)

//go:embed layout pages site stats auth error email
var Files embed.FS

var LoginForm = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"layout/flash.html",
		"layout/csrf.html",
		"layout/captcha.html",
		"auth/login_form.html",
	),
)

var RegisterForm = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"layout/flash.html",
		"layout/csrf.html",
		"layout/captcha.html",
		"auth/register_form.html",
	),
)

var Error = template.Must(template.ParseFS(Files,
	"error/error.html",
))

var ActivationEmail = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"email/activation_code.html",
	),
)

var Activate = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"layout/flash.html",
		"layout/csrf.html",
		"auth/activate.html",
	),
)

var Home = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
	),
)

var Sites = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"site/index.html",
		"layout/footer.html",
	),
)

var SiteNew = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"site/new.html",
	),
)

var Pricing = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"pages/pricing.html",
	),
)

var Markdown = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"pages/markdown.html",
	),
)

var AddSnippet = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"layout/flash.html",
		"layout/csrf.html",
		"site/snippet.html",
	),
)

var Stats = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"stats/stats.html",
	),
)

var WaitingFirstPageView = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"stats/waiting_first_pageview.html",
	),
)

var SiteLocked = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"stats/site_locked.html",
	),
)

type NewSite struct {
	IsFirstSite bool
	IsAtLimit   bool
	SiteLimit   int
}

type Errors struct {
	Status     int
	StatusText string
}

// The font used is Contrail One face 700-bold italic with size 150
const Logo = "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4KPHN2ZyB2aWV3Qm94PSItODM1Ljg2NCAtMTI4LjkzNiA5MDIuODYgMTg3LjA3NiIgd2lkdGg9IjkwMi44NiIgaGVpZ2h0PSIxODcuMDc2IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogIDxwYXRoIGQ9Ik0gLTgxNS4xMzQgMTAuOTM4IEwgLTgyMy42MyAtOTkuMzY1IFEgLTgyMy42MyAtMTAxLjA0OSAtODIxLjY1MyAtMTAxLjEyMyBMIC04MDYuNDE4IC0xMDEuNjM1IFEgLTgwNC44MDcgLTEwMS42MzUgLTgwNC44MDcgLTEwMC4yNDQgTCAtODAxLjM2NSAtMTcuNjI3IEwgLTc4MS45NTYgLTk5LjM2NSBRIC03ODEuNzM2IC0xMDAuMzE3IC03ODEuNDA2IC0xMDAuNjgzIFEgLTc4MS4wNzcgLTEwMS4wNDkgLTc4MC4xOTggLTEwMS4xMjMgTCAtNzY1LjAzNyAtMTAxLjcwOSBRIC03NjMuNDI1IC0xMDEuNzA5IC03NjMuNDI1IC0xMDAuMzE3IEwgLTc5NS41NzkgMTAuNDI1IFEgLTc5Ni4wOTEgMTIuMTgzIC03OTcuMzM2IDEyLjE4MyBMIC04MTMuNTIzIDEyLjc2OSBRIC04MTQuOTg4IDEyLjY5NiAtODE1LjEzNCAxMC45MzggWiIgc3R5bGU9ImZpbGw6IHJnYig1MSwgNTEsIDUxKTsgd2hpdGUtc3BhY2U6IHByZTsiLz4KICA8cGF0aCBkPSJNIC03NDEuMjQxIC04NC4wNTcgUSAtNzQxLjI0MSAtODQuMDU3IC03NDMuNTExIC04NC4wNTcgUSAtNzQ1Ljc4MiAtODQuMDU3IC03NDguNDkyIC04Ni45MTQgUSAtNzUxLjIwMiAtODkuNzcgLTc1MS4yMDIgLTk0LjI3NSBRIC03NTEuMjAyIC05OC43NzkgLTc0Ny41NzYgLTEwMi4zMzEgUSAtNzQzLjk1MSAtMTA1Ljg4MyAtNzM5LjQxIC0xMDUuODgzIFEgLTczNC44NjkgLTEwNS44ODMgLTczMi4wNDkgLTEwMy4wMjcgUSAtNzI5LjIyOSAtMTAwLjE3MSAtNzI5LjIyOSAtOTUuNjY2IFEgLTcyOS4yMjkgLTkxLjE2MiAtNzMyLjk2NCAtODcuNjEgUSAtNzM2LjcgLTg0LjA1NyAtNzQxLjI0MSAtODQuMDU3IFogTSAtNzMzLjExMSAtNzQuMTcgTCAtNzQyLjYzMiAxMC4yNzkgUSAtNzQyLjc3OSAxMS4zMDQgLTc0My4xNDUgMTEuNzQzIFEgLTc0My41MTEgMTIuMTgzIC03NDQuNTM2IDEyLjI1NiBMIC03NTkuMjU4IDEyLjc2OSBRIC03NjAuODcgMTIuNjIyIC03NjAuODcgMTEuMzc3IEwgLTc1MS4yMDIgLTczLjgwMyBRIC03NTAuOTA5IC03NS40MTUgLTc0OS4zNyAtNzUuNTYxIEwgLTczNS4wODggLTc2LjA3NCBRIC03MzMuMTExIC03NS45MjcgLTczMy4xMTEgLTc0LjE3IFoiIHN0eWxlPSJmaWxsOiByZ2IoNTEsIDUxLCA1MSk7IHdoaXRlLXNwYWNlOiBwcmU7Ii8+CiAgPHBhdGggZD0iTSAtNjkzLjA0NCAtNzQuMTcgTCAtNjkzLjg1IC02Ny40MzEgUSAtNjgzLjMwMyAtNzYuNzMzIC02NzMuNjM1IC03Ni43MzMgUSAtNjYxLjMzIC03Ni43MzMgLTY1OC4xMDggLTYzLjI1NyBRIC02NTcuMjI5IC01OS41OTQgLTY1Ny4yMjkgLTU2LjMzNSBRIC02NTcuMjI5IC01My4wNzYgLTY1Ny40NDggLTUxLjA5OCBMIC02NjQuNDggMTAuMjc5IFEgLTY2NC43NzMgMTIuMTEgLTY2Ni4zMTEgMTIuMjU2IEwgLTY4MS4xMDYgMTIuNzY5IFEgLTY4Mi43MTcgMTIuNjIyIC02ODIuNzE3IDExLjM3NyBMIC02NzYuMDUyIC00Ni42MzEgUSAtNjc1LjgzMiAtNDguNTM1IC02NzUuODMyIC01MC4wNzMgUSAtNjc1LjgzMiAtNTYuMDc5IC02NzguNjUyIC01OC4yMzkgUSAtNjgxLjQ3MiAtNjAuNCAtNjg1Ljc5MyAtNjAuNCBRIC02OTAuMTE0IC02MC40IC02OTQuODAyIC01OC41NjkgTCAtNzAyLjYzOSAxMC4yNzkgUSAtNzAyLjkzMiAxMi4xMSAtNzA0LjQ3IDEyLjI1NiBMIC03MTkuMjY1IDEyLjc2OSBRIC03MjAuODc2IDEyLjYyMiAtNzIwLjg3NiAxMS4zNzcgTCAtNzExLjIwOCAtNzMuODAzIFEgLTcxMC45MTUgLTc1LjQxNSAtNzA5LjM3NyAtNzUuNTYxIEwgLTY5NS4wMjIgLTc2LjA3NCBRIC02OTMuMDQ0IC03Ni4wMDEgLTY5My4wNDQgLTc0LjE3IFoiIHN0eWxlPSJmaWxsOiByZ2IoNTEsIDUxLCA1MSk7IHdoaXRlLXNwYWNlOiBwcmU7Ii8+CiAgPHBhdGggZD0iTSAtNTkwLjIxMSAtNzQuNDYzIFEgLTU4OC41OTkgLTczLjg3NyAtNTg4LjU5OSAtNzEuNjc5IEwgLTU5MC4yODQgLTU4LjQyMyBRIC01OTAuNDMgLTU2Ljk1OCAtNTkyLjI5OCAtNTYuOTU4IFEgLTU5NC4xNjYgLTU2Ljk1OCAtNTk1Ljg1IC01Ny42OSBRIC02MDIuOTU1IC02MC44NCAtNjA3LjY0MiAtNjAuODQgUSAtNjEyLjMzIC02MC44NCAtNjE1LjExMyAtNTYuOTU4IFEgLTYxNy44OTYgLTUzLjA3NiAtNjE5LjU4MSAtNDUuODI1IFEgLTYyMi43MyAtMzIuMjc1IC02MjIuNzMgLTIxLjA2OSBRIC02MjIuNzMgLTMuMDUxIC02MTMuNDI4IC0zLjA1MSBRIC02MDYuNjkgLTMuMDUxIC02MDIuNTE1IC00Ljc3MyBRIC01OTguMzQxIC02LjQ5NCAtNTk3LjA5NSAtNi40OTQgUSAtNTk0LjUzMiAtNi40OTQgLTU5NC41MzIgLTQuMjk3IEwgLTU5Ni4wNyA4LjU5NCBRIC01OTYuMjE2IDkuOTEyIC01OTYuNjE5IDEwLjE2OSBRIC01OTcuMDIyIDEwLjQyNSAtNTk3LjY4MSAxMC43MTggUSAtNjA2LjU0NCAxMy4yODIgLTYxNC4zNDQgMTMuMjgyIFEgLTYyMi4xNDQgMTMuMjgyIC02MjYuNjQ5IDExLjMwNCBRIC02MzEuMTUzIDkuMzI2IC02MzMuNzkgNi4yMTQgUSAtNjM2LjQyNiAzLjEwMSAtNjM4LjAzOCAtMS4yOTQgUSAtNjQwLjg5NCAtOC44MzggLTY0MC44OTQgLTE5Ljg2MSBRIC02NDAuODk0IC0zMC44ODMgLTYzOS40NjYgLTM5LjA1IFEgLTYzOC4wMzggLTQ3LjIxNiAtNjM2LjQ2MyAtNTIuMzQzIFEgLTYzNC44ODggLTU3LjQ3IC02MzIuMjUyIC02Mi4xOTUgUSAtNjI5LjYxNSAtNjYuOTE5IC02MjYuMDk5IC02OS45OTUgUSAtNjE4LjQwOSAtNzYuNzMzIC02MDUuNzM4IC03Ni43MzMgUSAtNTk3LjI0MiAtNzYuNzMzIC01OTAuMjExIC03NC40NjMgWiIgc3R5bGU9ImZpbGw6IHJnYig1MSwgNTEsIDUxKTsgd2hpdGUtc3BhY2U6IHByZTsiLz4KICA8cGF0aCBkPSJNIC01NzMuMDM1IC0xLjI1NyBRIC01NzMuMDM1IC0xLjI1NyAtNTczLjg5NiAtMy40MzYgUSAtNTc0Ljc1NyAtNS42MTUgLTU3NS4zNzkgLTkuOSBRIC01NzYuMDAyIC0xNC4xODQgLTU3Ni4wMDIgLTIxLjY1NSBRIC01NzYuMDAyIC0yOS4xMjYgLTU3NC4zOSAtNDAuMTEyIFEgLTU2OS4wNDQgLTc2LjczMyAtNTQxLjEzOCAtNzYuNzMzIFEgLTUzMC4zNzIgLTc2LjczMyAtNTI0LjY1OSAtNjkuMzM2IFEgLTUxOC45NDYgLTYxLjkzOCAtNTE4Ljk0NiAtNTAuMTQ2IFEgLTUxOC45NDYgLTM4LjM1NCAtNTIwLjQxMSAtMjcuMjIxIFEgLTUyMC42MzEgLTI1LjI0NCAtNTIyLjY4MSAtMjUuMTcxIEwgLTU1OC44NjMgLTIzLjkyNSBRIC01NTguOTM2IC0yMi44MjcgLTU1OC45MzYgLTIwLjk5NiBRIC01NTguOTM2IC0xMi43MTkgLTU1Ni4wNDMgLTcuODg1IFEgLTU1My4xNSAtMy4wNTEgLTU0Ny4yOTEgLTMuMDUxIFEgLTU0MS40MzEgLTMuMDUxIC01MzYuNDE0IC00Ljc3MyBRIC01MzEuMzk3IC02LjQ5NCAtNTMwLjUxOCAtNi40OTQgUSAtNTI5LjYzOSAtNi40OTQgLTUyOC45OCAtNS45MDggUSAtNTI4LjMyMSAtNS4zMjIgLTUyOC4zMjEgLTQuMjk3IEwgLTUyOS43ODYgOC41OTQgUSAtNTI5Ljg1OSA5LjY5MyAtNTMwLjE4OSAxMC4wMjIgUSAtNTMwLjUxOCAxMC4zNTIgLTUzMS40NyAxMC43MTggUSAtNTQwLjMzMyAxMy4yODIgLTU0OC41MzYgMTMuMjgyIFEgLTU1Ni43MzkgMTMuMjgyIC01NjEuMjggMTEuMzA0IFEgLTU2NS44MjEgOS4zMjYgLTU2OC41NjggNi4yMTQgUSAtNTcxLjMxNCAzLjEwMSAtNTczLjAzNSAtMS4yNTcgWiBNIC01NTcuMTA1IC00Mi4zMDkgTCAtNTU3Ljc2NCAtMzcuNjIyIEwgLTUzOC42NDggLTM4LjU3NCBRIC01MzcuNDAzIC0zOC42NDcgLTUzNy4wMzcgLTM5LjAxMyBRIC01MzUuOTM4IC00MC4xMTIgLTUzNS45MzggLTQ1LjA1NiBRIC01MzUuOTM4IC01MCAtNTM2LjI2OCAtNTIuMDE0IFEgLTUzNi41OTcgLTU0LjAyOCAtNTM3LjQwMyAtNTYuMTUyIFEgLTUzOS4wODggLTYwLjg0IC01NDMuMTUzIC02MC44NCBRIC01NDcuMjE3IC02MC44NCAtNTQ5LjU2MSAtNTkuNDQ4IFEgLTU1NC43NjEgLTU2LjIyNSAtNTU3LjEwNSAtNDIuMzA5IFoiIHN0eWxlPSJmaWxsOiByZ2IoNTEsIDUxLCA1MSk7IHdoaXRlLXNwYWNlOiBwcmU7Ii8+CiAgPHBhdGggZD0iTSAtNDI3LjgzNyAtOTkuOTUxIEwgLTQxOS4yNjggMTAuMzUyIFEgLTQxOS4yNjggMTIuMDM2IC00MjEuMTcyIDEyLjE4MyBMIC00MzYuNDggMTIuNzY5IFEgLTQzOC4wMTggMTIuNzY5IC00MzguMDE4IDExLjM3NyBMIC00MzkuMTE3IC0xMi43OTMgTCAtNDU1Ljc0MyAtMTIuMTMzIEwgLTQ2MS4yMzYgMTAuNDI1IFEgLTQ2MS40NTYgMTEuNDUxIC00NjEuODIyIDExLjgxNyBRIC00NjIuMTg4IDEyLjE4MyAtNDYzLjE0IDEyLjI1NiBMIC00NzguMzAxIDEyLjc2OSBRIC00NzkuNzY2IDEyLjc2OSAtNDc5Ljc2NiAxMS4zNzcgTCAtNDQ3LjYxMyAtOTkuNDM4IFEgLTQ0Ny4xIC0xMDEuMDQ5IC00NDUuNjM1IC0xMDEuMTIzIEwgLTQyOS41MjIgLTEwMS43MDkgUSAtNDI3Ljk4NCAtMTAxLjYzNSAtNDI3LjgzNyAtOTkuOTUxIFogTSAtNDQxLjY4IC02OS43NzUgTCAtNDUxLjkzNCAtMjcuNzM0IEwgLTQzOS43NzYgLTI4LjE3NCBaIiBzdHlsZT0iZmlsbDogcmdiKDUxLCA1MSwgNTEpOyB3aGl0ZS1zcGFjZTogcHJlOyIvPgogIDxwYXRoIGQ9Ik0gLTM3NS42ODMgLTc0LjE3IEwgLTM3Ni40ODkgLTY3LjQzMSBRIC0zNjUuOTQyIC03Ni43MzMgLTM1Ni4yNzQgLTc2LjczMyBRIC0zNDMuOTY5IC03Ni43MzMgLTM0MC43NDYgLTYzLjI1NyBRIC0zMzkuODY3IC01OS41OTQgLTMzOS44NjcgLTU2LjMzNSBRIC0zMzkuODY3IC01My4wNzYgLTM0MC4wODcgLTUxLjA5OCBMIC0zNDcuMTE4IDEwLjI3OSBRIC0zNDcuNDExIDEyLjExIC0zNDguOTQ5IDEyLjI1NiBMIC0zNjMuNzQ0IDEyLjc2OSBRIC0zNjUuMzU2IDEyLjYyMiAtMzY1LjM1NiAxMS4zNzcgTCAtMzU4LjY5MSAtNDYuNjMxIFEgLTM1OC40NzEgLTQ4LjUzNSAtMzU4LjQ3MSAtNTAuMDczIFEgLTM1OC40NzEgLTU2LjA3OSAtMzYxLjI5MSAtNTguMjM5IFEgLTM2NC4xMTEgLTYwLjQgLTM2OC40MzIgLTYwLjQgUSAtMzcyLjc1MyAtNjAuNCAtMzc3LjQ0MSAtNTguNTY5IEwgLTM4NS4yNzggMTAuMjc5IFEgLTM4NS41NzEgMTIuMTEgLTM4Ny4xMDkgMTIuMjU2IEwgLTQwMS45MDQgMTIuNzY5IFEgLTQwMy41MTUgMTIuNjIyIC00MDMuNTE1IDExLjM3NyBMIC0zOTMuODQ3IC03My44MDMgUSAtMzkzLjU1NCAtNzUuNDE1IC0zOTIuMDE2IC03NS41NjEgTCAtMzc3LjY2IC03Ni4wNzQgUSAtMzc1LjY4MyAtNzYuMDAxIC0zNzUuNjgzIC03NC4xNyBaIiBzdHlsZT0iZmlsbDogcmdiKDUxLCA1MSwgNTEpOyB3aGl0ZS1zcGFjZTogcHJlOyIvPgogIDxwYXRoIGQ9Ik0gLTI2MC40MDYgLTc0LjE3IEwgLTI2OS45MjggMTAuMjc5IFEgLTI3MC4xNDcgMTIuMTgzIC0yNzEuODMyIDEyLjI1NiBMIC0yODYuNjI3IDEyLjc2OSBRIC0yODguMTY1IDEyLjYyMiAtMjg4LjE2NSAxMS4zNzcgTCAtMjg2Ljc3MyAxLjQxNiBRIC0yOTAuMTQyIDcuMTI5IC0yOTQuMTM0IDEwLjIwNSBRIC0yOTguMTI2IDEzLjI4MiAtMzAzLjEwNiAxMy4yODIgUSAtMzA4LjA4NyAxMy4yODIgLTMxMS42MDIgMTEuNTYgUSAtMzE1LjExOCA5LjgzOSAtMzE3LjM1MiA3LjAxOSBRIC0zMTkuNTg2IDQuMiAtMzIwLjk3NyAwLjI0NCBRIC0zMjMuNTQxIC02Ljc4NyAtMzIzLjU0MSAtMTguOTgyIFEgLTMyMy41NDEgLTMxLjE3NiAtMzIxLjg5MyAtMzkuNjczIFEgLTMyMC4yNDUgLTQ4LjE2OSAtMzE5IC01Mi42IFEgLTMxNy43NTUgLTU3LjAzMSAtMzE1LjY2NyAtNjEuNzE4IFEgLTMxMy41OCAtNjYuNDA2IC0zMTAuOTQzIC02OS41NTUgUSAtMzA0LjkzNyAtNzYuNzMzIC0yOTUuNzgyIC03Ni43MzMgUSAtMjg4Ljk3MSAtNzYuNzMzIC0yODUuMzgyIC03My40NzQgUSAtMjgxLjc5MyAtNzAuMjE1IC0yNzkuMzc2IC02My44NDIgTCAtMjc4LjQ5NyAtNzMuODAzIFEgLTI3OC4zNSAtNzUuNDE1IC0yNzYuNjY2IC03NS41NjEgTCAtMjYyLjMxIC03Ni4wNzQgUSAtMjYwLjQwNiAtNzUuOTI3IC0yNjAuNDA2IC03NC4xNyBaIE0gLTMwNS4wODQgLTIxLjA2OSBRIC0zMDUuMDg0IC0zLjU2NCAtMjk1LjcwOSAtMy41NjQgUSAtMjkxLjYwNyAtMy41NjQgLTI4OS4wNDQgLTcuMzM2IFEgLTI4Ni40OCAtMTEuMTA4IC0yODQuNjEzIC0yMi40NjEgUSAtMjgyLjc0NSAtMzMuODEzIC0yODIuNzQ1IC0zOS4wMTMgUSAtMjgyLjc0NSAtNDQuMjE0IC0yODIuODE4IC00NS42MDUgUSAtMjgyLjg5MSAtNDYuOTk3IC0yODMuMTExIC00OS4zMDQgUSAtMjgzLjMzMSAtNTEuNjExIC0yODMuOTE3IC01My4xODYgUSAtMjg0LjUwMyAtNTQuNzYgLTI4NS40NTUgLTU2LjM3MiBRIC0yODcuMzU5IC01OS41MjEgLTI5MS42MDcgLTU5LjUyMSBRIC0yOTUuODU1IC01OS41MjEgLTI5OC4zNDYgLTU1Ljc0OSBRIC0zMDAuODM2IC01MS45NzcgLTMwMi4yMjcgLTQ1LjY3OCBRIC0zMDUuMDg0IC0zMi4yMDIgLTMwNS4wODQgLTIxLjA2OSBaIiBzdHlsZT0iZmlsbDogcmdiKDUxLCA1MSwgNTEpOyB3aGl0ZS1zcGFjZTogcHJlOyIvPgogIDxwYXRoIGQ9Ik0gLTIxNi40NiAtMTA5LjQ3MiBMIC0yMjkuOTM3IDEwLjIwNSBRIC0yMzAuMDgzIDExLjIzMSAtMjMwLjQ1IDExLjY3IFEgLTIzMC44MTYgMTIuMTEgLTIzMS44NDEgMTIuMTgzIEwgLTI0Ni41NjMgMTIuNzY5IFEgLTI0OC4xNzQgMTIuNjIyIC0yNDguMTc0IDExLjM3NyBMIC0yMzQuNjI0IC0xMDguOTYgUSAtMjM0LjQ3OCAtMTA5Ljk4NSAtMjM0LjExMiAtMTEwLjQyNCBRIC0yMzMuNzQ2IC0xMTAuODY0IC0yMzIuNzIgLTExMC45MzcgTCAtMjE4LjM2NSAtMTExLjQ1IFEgLTIxNi40NiAtMTExLjIzIC0yMTYuNDYgLTEwOS40NzIgWiIgc3R5bGU9ImZpbGw6IHJnYig1MSwgNTEsIDUxKTsgd2hpdGUtc3BhY2U6IHByZTsiLz4KICA8cGF0aCBkPSJNIC0xNDQuOTc4IC03NC4xNyBMIC0xODcuMjM4IDQ1LjcyOCBRIC0xODcuODI0IDQ3LjU1OSAtMTg5LjM2MiA0Ny42MzIgTCAtMjA0LjA4NCA0OC4xNDUgUSAtMjA1LjY5NSA0OC4wNzIgLTIwNS42OTUgNDYuNjggTCAtMTkzLjA5OCAxMS4xNTggTCAtMjA0LjUyNCAtNzMuODAzIFEgLTIwNC41MjQgLTc1LjQ4OCAtMjAyLjY5MyAtNzUuNTYxIEwgLTE4OC4yNjQgLTc2LjA3NCBRIC0xODYuNTc5IC03Ni4wMDEgLTE4Ni40MzMgLTc0LjE3IEwgLTE4MC44NjYgLTEyLjU3MyBMIC0xNjMuMDY5IC03My44MDMgUSAtMTYyLjYyOSAtNzUuNDg4IC0xNjEuMTY0IC03NS41NjEgTCAtMTQ2LjgwOSAtNzYuMDc0IFEgLTE0NC45NzggLTc1LjkyNyAtMTQ0Ljk3OCAtNzQuMTcgWiIgc3R5bGU9ImZpbGw6IHJnYig1MSwgNTEsIDUxKTsgd2hpdGUtc3BhY2U6IHByZTsiLz4KICA8cGF0aCBkPSJNIC0xMjcuODMzIC03NS43ODEgUSAtMTI2LjE0OCAtOTEuODk0IC0xMjUuNjM2IC05Mi40NDQgUSAtMTI1LjEyMyAtOTIuOTkzIC0xMjQuMDk4IC05My4wNjYgTCAtMTEwLjEwOCAtOTMuNTc5IFEgLTEwOC4xMzEgLTkzLjQzMiAtMTA4LjEzMSAtOTEuNzQ4IFEgLTEwOS41MjIgLTgzLjYxOCAtMTEwLjQ3NSAtNzYuMTQ3IEwgLTkzLjQwOSAtNzYuNzMzIFEgLTkxLjY1MSAtNzYuNjYgLTkxLjY1MSAtNzQuODI5IEwgLTkzLjcwMiAtNjMuMTgzIFEgLTkzLjk5NSAtNjEuNDI1IC05NS41MzMgLTYxLjM1MiBMIC0xMTIuMzA2IC02MC42OTMgUSAtMTE3LjI4NiAtMTUuODY5IC0xMTcuMjg2IC0xMS4yNTUgUSAtMTE3LjI4NiAtMi42ODUgLTEwOS45NjIgLTIuNjg1IFEgLTEwNi4yMjcgLTIuNjg1IC0xMDIuNzExIC0zLjg1NyBRIC05OS4xOTUgLTUuMDI5IC05OC4zMTYgLTUuMDI5IFEgLTk2Ljc3OCAtNS4wMjkgLTk2Ljc3OCAtMy4xMjUgTCAtOTguMzE2IDkuNjkzIFEgLTk4LjQ2MyAxMC43OTEgLTk4Ljg2NiAxMS4zMDQgUSAtOTkuMjY5IDExLjgxNyAtMTAyLjQ1NSAxMi41NDkgUSAtMTA1LjY0MSAxMy4yODIgLTExMy4zMzEgMTMuMjgyIFEgLTEyNC42MSAxMy4yODIgLTEzMC4wMyA4LjA0NSBRIC0xMzUuNDUgMi44MDggLTEzNS40NSAtNy41OTIgUSAtMTM1LjQ1IC0xNy45OTMgLTEyOS42NjQgLTYwLjQ3MyBMIC0xMzYuODQyIC02MC4xOCBRIC0xMzguNjczIC02MC4yNTQgLTEzOC42NzMgLTYyLjAxMSBMIC0xMzYuNjk1IC03My44MDMgUSAtMTM2LjQ3NiAtNzUuNDg4IC0xMzQuNzkxIC03NS41NjEgWiIgc3R5bGU9ImZpbGw6IHJnYig1MSwgNTEsIDUxKTsgd2hpdGUtc3BhY2U6IHByZTsiLz4KICA8cGF0aCBkPSJNIC02NS4yMTIgLTg0LjA1NyBRIC02NS4yMTIgLTg0LjA1NyAtNjcuNDgzIC04NC4wNTcgUSAtNjkuNzUzIC04NC4wNTcgLTcyLjQ2MyAtODYuOTE0IFEgLTc1LjE3MyAtODkuNzcgLTc1LjE3MyAtOTQuMjc1IFEgLTc1LjE3MyAtOTguNzc5IC03MS41NDggLTEwMi4zMzEgUSAtNjcuOTIyIC0xMDUuODgzIC02My4zODEgLTEwNS44ODMgUSAtNTguODQgLTEwNS44ODMgLTU2LjAyIC0xMDMuMDI3IFEgLTUzLjIwMSAtMTAwLjE3MSAtNTMuMjAxIC05NS42NjYgUSAtNTMuMjAxIC05MS4xNjIgLTU2LjkzNiAtODcuNjEgUSAtNjAuNjcxIC04NC4wNTcgLTY1LjIxMiAtODQuMDU3IFogTSAtNTcuMDgyIC03NC4xNyBMIC02Ni42MDQgMTAuMjc5IFEgLTY2Ljc1IDExLjMwNCAtNjcuMTE3IDExLjc0MyBRIC02Ny40ODMgMTIuMTgzIC02OC41MDggMTIuMjU2IEwgLTgzLjIzIDEyLjc2OSBRIC04NC44NDEgMTIuNjIyIC04NC44NDEgMTEuMzc3IEwgLTc1LjE3MyAtNzMuODAzIFEgLTc0Ljg4IC03NS40MTUgLTczLjM0MiAtNzUuNTYxIEwgLTU5LjA2IC03Ni4wNzQgUSAtNTcuMDgyIC03NS45MjcgLTU3LjA4MiAtNzQuMTcgWiIgc3R5bGU9ImZpbGw6IHJnYig1MSwgNTEsIDUxKTsgd2hpdGUtc3BhY2U6IHByZTsiLz4KICA8cGF0aCBkPSJNIDcuNjU5IC03NC40NjMgUSA5LjI3IC03My44NzcgOS4yNyAtNzEuNjc5IEwgNy41ODYgLTU4LjQyMyBRIDcuNDM5IC01Ni45NTggNS41NzIgLTU2Ljk1OCBRIDMuNzA0IC01Ni45NTggMi4wMTkgLTU3LjY5IFEgLTUuMDg1IC02MC44NCAtOS43NzMgLTYwLjg0IFEgLTE0LjQ2IC02MC44NCAtMTcuMjQzIC01Ni45NTggUSAtMjAuMDI2IC01My4wNzYgLTIxLjcxMSAtNDUuODI1IFEgLTI0Ljg2IC0zMi4yNzUgLTI0Ljg2IC0yMS4wNjkgUSAtMjQuODYgLTMuMDUxIC0xNS41NTkgLTMuMDUxIFEgLTguODIgLTMuMDUxIC00LjY0NiAtNC43NzMgUSAtMC40NzEgLTYuNDk0IDAuNzc0IC02LjQ5NCBRIDMuMzM4IC02LjQ5NCAzLjMzOCAtNC4yOTcgTCAxLjggOC41OTQgUSAxLjY1MyA5LjkxMiAxLjI1IDEwLjE2OSBRIDAuODQ4IDEwLjQyNSAwLjE4OCAxMC43MTggUSAtOC42NzQgMTMuMjgyIC0xNi40NzQgMTMuMjgyIFEgLTI0LjI3NCAxMy4yODIgLTI4Ljc3OSAxMS4zMDQgUSAtMzMuMjgzIDkuMzI2IC0zNS45MiA2LjIxNCBRIC0zOC41NTcgMy4xMDEgLTQwLjE2OCAtMS4yOTQgUSAtNDMuMDI0IC04LjgzOCAtNDMuMDI0IC0xOS44NjEgUSAtNDMuMDI0IC0zMC44ODMgLTQxLjU5NiAtMzkuMDUgUSAtNDAuMTY4IC00Ny4yMTYgLTM4LjU5MyAtNTIuMzQzIFEgLTM3LjAxOSAtNTcuNDcgLTM0LjM4MiAtNjIuMTk1IFEgLTMxLjc0NSAtNjYuOTE5IC0yOC4yMyAtNjkuOTk1IFEgLTIwLjUzOSAtNzYuNzMzIC03Ljg2OCAtNzYuNzMzIFEgMC42MjggLTc2LjczMyA3LjY1OSAtNzQuNDYzIFoiIHN0eWxlPSJmaWxsOiByZ2IoNTEsIDUxLCA1MSk7IHdoaXRlLXNwYWNlOiBwcmU7Ii8+CiAgPHBhdGggZD0iTSA2MC4xMDEgLTU5LjAwOCBMIDUxLjgyNCAtNTkuMzc1IFEgNDguMzA5IC01OS4zNzUgNDYuMjk0IC01Ny43NjMgUSA0NC4yOCAtNTYuMTUyIDQ0LjI4IC01My4xMTIgUSA0NC4yOCAtNTAuMDczIDQ1LjM3OSAtNDYuNDExIFEgNDYuODQ0IC00MS41MDQgNTAuNDY5IC0zNC45ODUgUSA1NC4wOTUgLTI4LjQ2NiA1NS40ODYgLTI1Ljc1NyBRIDU5LjczNCAtMTcuMzM0IDU5LjczNCAtOS43MTYgUSA1OS43MzQgMS4zNDMgNTMuMjUyIDcuMzEyIFEgNDYuNzcgMTMuMjgyIDMzLjU4NyAxMy4yODIgUSAyNy4wNjggMTMuMjgyIDIzLjE4NiAxMi41ODYgUSAxOS4zMDUgMTEuODkgMTguOTAyIDExLjQ4NyBRIDE4LjQ5OSAxMS4wODQgMTguNDk5IDEwLjIwNSBMIDIwLjQwMyAtMy4zNDQgUSAyMC41NSAtNS4xNzUgMjEuNjg1IC01LjE3NSBRIDIyLjgyIC01LjE3NSAyNi4wNDMgLTQuMzcgUSAyOS4yNjYgLTMuNTY0IDMxLjY4MyAtMy41NjQgUSA0MC42MTggLTMuNTY0IDQwLjYxOCAtMTAuNjY5IFEgNDAuNjE4IC0xNS43MjIgMzguNDU3IC0yMC4xMTcgUSAzNi4yOTcgLTI0LjUxMSAzMy43NyAtMjkuMTk5IFEgMzEuMjQzIC0zMy44ODYgMjkuODUyIC0zNi43MDYgUSAyOC40NiAtMzkuNTI2IDI2Ljk1OCAtNDQuNTA3IFEgMjUuNDU3IC00OS40ODcgMjUuNDU3IC01My44ODIgUSAyNS40NTcgLTY0LjEzNSAzMS41NzMgLTcwLjQzNCBRIDM3LjY4OCAtNzYuNzMzIDQ5LjQwNyAtNzYuNzMzIFEgNTYuOTUxIC03Ni43MzMgNjEuMTk5IC03Ni4wMDEgUSA2My4wMyAtNzUuNzA4IDYzLjAzIC03NC4wOTYgTCA2MS40MTkgLTYxLjA1OSBRIDYxLjQxOSAtNTkuMDA4IDYwLjEwMSAtNTkuMDA4IFoiIHN0eWxlPSJmaWxsOiByZ2IoNTEsIDUxLCA1MSk7IHdoaXRlLXNwYWNlOiBwcmU7Ii8+Cjwvc3ZnPg=="

type Context struct {
	Title         string
	CurrentUser   *models.User
	Data          map[string]any
	CSRF          template.HTML
	Captcha       template.HTMLAttr
	Errors        map[string]string
	Form          url.Values
	Code          uint64
	Config        *config.Config
	HasInvitation bool
	HasPin        bool
	Flash         *flash.Flash
	NewSite       *NewSite
	Error         *Errors
	Position      int
	Page          string
	Content       template.HTML
	ModTime       time.Time
	Site          *models.Site
	IsFIrstSite   bool
	SitesOverview []models.SiteOverView
	EmailReport   bool
	HasGoals      bool
	Owner         *models.User
}

func New(ctx context.Context, f ...func(c *Context)) *Context {
	c := &Context{
		Data:        make(map[string]any),
		CSRF:        getCsrf(ctx),
		Captcha:     getCaptcha(ctx),
		CurrentUser: models.GetUser(ctx),
		Config:      config.Get(ctx),
		Code:        GetActivationCode(ctx),
		Flash:       flash.Get(ctx),
	}
	if len(f) > 0 {
		f[0](c)
	}
	return c
}

type csrfTokenCtxKey struct{}

func getCsrf(ctx context.Context) template.HTML {
	if c := ctx.Value(csrfTokenCtxKey{}); c != nil {
		return c.(template.HTML)
	}
	return template.HTML("")
}

type captchaTokenKey struct{}

func SetCaptcha(ctx context.Context, x template.HTMLAttr) context.Context {
	return context.WithValue(ctx, captchaTokenKey{}, x)
}

func SetCSRF(ctx context.Context, x template.HTML) context.Context {
	return context.WithValue(ctx, csrfTokenCtxKey{}, x)
}

func getCaptcha(ctx context.Context) template.HTMLAttr {
	if c := ctx.Value(captchaTokenKey{}); c != nil {
		return c.(template.HTMLAttr)
	}
	return template.HTMLAttr("")
}

func (t *Context) VinceURL() template.HTML {
	return template.HTML("http://localhost:8080")
}

func (t *Context) Logo(width, height int) template.HTML {
	return template.HTML(fmt.Sprintf(
		`<img alt="Vince Analytics logo" width=%d height=%d src=%q>`,
		width, height, Logo,
	))
}

func (t *Context) SetPosition(pos int) *Context {
	t.Position = pos
	return t
}

func (t *Context) OnboardSteps() []string {
	return []string{
		"Register", "Activate account", "Add site info", "Install snippet",
	}
}

func (t *Context) Validate(name string) template.HTML {
	if t.Errors != nil {
		o, _ := octicon.Icon("alert-fill", 12)
		if v, ok := t.Errors[name]; ok {
			return template.HTML(fmt.Sprintf(`
<div class="FormControl-inlineValidation">
    %s
    <span>%s</span>
</div>
		`, o, v))
		}
	}
	return template.HTML("")
}

func (t *Context) Icon(name string, height int, class ...string) (template.HTML, error) {
	return octicon.IconTemplateFunc(name, height, class...)
}

func (t *Context) InputField(name string) template.HTMLAttr {
	var s strings.Builder
	if t.Errors != nil && t.Errors[name] != "" {
		s.WriteString(`invalid="true"`)
	}
	if t.Form != nil && t.Form.Get(name) != "" {
		s.WriteString(fmt.Sprintf("value=%q", t.Form.Get(name)))
	}
	return template.HTMLAttr(s.String())
}

var steps = []string{
	"Register", "Activate account", "Add site info", "Install snippet",
}

func (t *Context) CurrentStep() string {
	return steps[t.Position]
}

func (t *Context) StepsBefore() []string {
	return steps[0:t.Position]
}

func (t *Context) StepsAfter() []string {
	return steps[t.Position:]
}

type activationCodeKey struct{}

func SetActivationCode(ctx context.Context, code uint64) context.Context {
	return context.WithValue(ctx, activationCodeKey{}, code)
}

func GetActivationCode(ctx context.Context) uint64 {
	if v := ctx.Value(activationCodeKey{}); v != nil {
		return v.(uint64)
	}
	return 0
}

func (t *Context) Format(n uint64) string {
	switch {
	case n >= 1_000 && n < 1_000_000:
		thousands := (n / 100) / 10
		return fmt.Sprintf("%dK", thousands)
	case n >= 1_000_000 && n < 1_000_000_000:
		millions := (n / 100_000) / 10
		return fmt.Sprintf("%dM", millions)
	case n >= 1_000_000_000 && n < 1_000_000_000_000:
		billions := (n / 100_000_000) / 10
		return fmt.Sprintf("%dB", billions)
	default:
		return strconv.FormatUint(n, 10)
	}
}

func (t *Context) Plans() []plans.Plan {
	return plans.All
}

func (t *Context) Snippet() template.HTML {
	track := fmt.Sprintf("https://%s/js/script.js", t.Config.Url)
	if t.Site.CustomDomain != nil {
		track = fmt.Sprintf("https://%s/js/index.js", t.Site.CustomDomain.Domain)
	}
	src := fmt.Sprintf("<script defer data-domain=%q src=%q></script>", t.Site.Domain, track)
	return template.HTML(src)
}
