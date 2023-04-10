package models

import "html/template"

type SiteOverView struct {
	// svg of the template spark lines
	SparkLine template.HTML

	Site *Site
}
