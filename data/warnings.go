package data

type Warning struct {
	Message      string
	TemplateLine SourceMarker
}

type Warnings []Warning
