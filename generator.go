package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"
)

type Entry struct {
	URL        string
	Author     string
	Year       int
	Complaints []string
}

func main() {
	entries := []*Entry{}
	entryBytes, err := ioutil.ReadFile("entries.json")
	if err != nil {
		fmt.Println("Error reading entries.json:", err)
		fmt.Println(string(entryBytes))
		os.Exit(1)
	}

	err = json.Unmarshal(entryBytes, &entries)
	if err != nil {
		fmt.Println("Error unmarshalling entries:", err)
		fmt.Println(string(entryBytes))
		os.Exit(1)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Year < entries[j].Year })

	entryTmpl, err := template.New("entry").Parse(`
# The List

{{ range . }}+ {{ .URL }} ({{ .Author }}, {{ .Year }})
{{ range .Complaints }}  - {{ . }}
{{ end }}{{ end }}
`)
	if err != nil {
		fmt.Println("Error parsing entry template:", err)
		os.Exit(1)
	}
	complaintMap := map[string][]*Entry{}
	for _, e := range entries {
		for _, c := range e.Complaints {
			complaintMap[c] = append(complaintMap[c], e)
		}
	}

	complaintTmpl, err := template.New("complaint").Funcs(template.FuncMap{"anchorID": anchorID}).Parse(`
# Reverse complaints index

It's a reverse complaints index, generated by https://github.com/ksimka/go-is-not-good/blob/master/generator.go (thanks to @capoferro)

{{ range $complaint, $entries := . }}+ <a id='{{ anchorID $complaint }}'>{{ $complaint }}</a>
{{ range $entries }}  - {{ .URL }} ({{ .Author }} {{ .Year }})
{{ end }}{{ end }}
`)
	if err != nil {
		fmt.Println("Error parsing complaint template:", err)
		os.Exit(1)
	}

	f, err := os.Create("README.md")
	if err != nil {
		fmt.Println("Error opening README.md:", err)
		os.Exit(1)
	}
	defer f.Close()

	err = copyContents("HEADER.md", f)
	entryTmpl.Execute(f, entries)
	complaintTmpl.Execute(f, complaintMap)
	err = copyContents("FOOTER.md", f)

}

func copyContents(origin string, target *os.File) error {
	f, err := os.Open(origin)
	if err != nil {
		fmt.Println("Error opening "+origin+":", err)
		os.Exit(1)
	}
	defer f.Close()

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	_, err = target.Write(contents)
	if err != nil {
		return err
	}

	return nil
}

// anchorID takes a string and returns an equivalent string suitable for use
// as an HTML anchor ID, with quotes removed and spaces replaced by hyphens.
// We also do some convenience replacements for making nice links, such as
// "C#" -> "csharp", ":=" -> "colon-equals", etc.
func anchorID(input string) string {
	replacer := strings.NewReplacer(`"`, "", "'", "", " ", "-", "`", "", "(", "", ")", "", "/", "-", "#", "sharp", ":=", "colon-equals")
	return url.QueryEscape(replacer.Replace(strings.ToLower(input)))
}
