package cmd

const subCommandsTpl = `
Available commands:
-------------------

{{ range $key, $c :=  .}} {{ $key }}: {{ $c.Short }}
{{ end }}
`

const usageTpl = `
Usage: xbee {{ .Path }} [OPTIONS] [ARG...]

{{ .Long }}

Aliases: {{ if .Aliases }}{{ range .Aliases }} {{ . }} {{ end }}{{ else }}None{{ end }}

ProcessOptions: {{ if .HasOptions }}{{ .OptionsToDisplay }}{{ else }}None{{ end }}

Global ProcessOptions: {{ .GlobalOptionsToDisplay }}
`
