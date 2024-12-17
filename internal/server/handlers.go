package server

import (
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/russross/blackfriday/v2"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func documentationHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("README.md")
	if err != nil {
		http.Error(w, "Documentation not found", http.StatusNotFound)
		return
	}

	htmlContent := blackfriday.Run(content)

	tmpl, err := template.New("doc").Parse(`
        <html>
        <head>
            <title>Bot Documentation</title>
        </head>
        <body>{{ .Content }}</body>
        </html>
    `)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Content template.HTML
	}{
		Content: template.HTML(htmlContent),
	}

	tmpl.Execute(w, data)
}
