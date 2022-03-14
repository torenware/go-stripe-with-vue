package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type templateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	Flash           string
	Warning         string
	Error           string
	IsAuthenticated int
	API             string
	CSSVersion      string
}

var functions = template.FuncMap{
	"formatCurrency": formatCurrency,
	"rfcDate":        formatDate,
}

// Formatter for currency. Unlike the course, I do not prepend the
// currency symbol, since I will also use this for raw float strings.
func formatCurrency(n int) string {
	f := float32(n) / float32(100)
	return fmt.Sprintf("%.2f", f)
}

func formatDate(date time.Time) string {
	formatted := date.Format(time.RFC822)
	return formatted
}

// Embed the templates into the binary. This is a go 1.16 feature.
// the "go:embed" command should be immediately after the double slashes,
// w/o a space before it.

//go:embed templates
var templateFS embed.FS

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	td.StringMap = make(map[string]string)
	td.StringMap["STRIPE_KEY"] = app.config.stripe.key
	td.StringMap["STRIPE_SECRET"] = app.config.stripe.secret
	td.API = app.config.api

	if session.Exists(r.Context(), "userID") {
		td.IsAuthenticated = 1
	}

	return td
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template
	var err error

	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", page)

	// Use the cache in production, but not in development
	_, templateInMap := app.templateCache[templateToRender]
	if app.config.env == "production" && templateInMap {
		// use cache
		t = app.templateCache[templateToRender]
	} else {
		// build the template
		t, err = app.parseTemplate(partials, page, templateToRender)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}
	if td == nil {
		td = &templateData{}
	}

	td = app.addDefaultData(td, r)

	err = t.Execute(w, td)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

func (app *application) parseTemplate(partials []string, page string, templateToRender string) (*template.Template, error) {
	var t *template.Template
	var err error

	// build partials
	if len(partials) > 0 {
		for i, x := range partials {
			partials[i] = fmt.Sprintf("templates/%s.partial.gohtml", x)
		}
	}
	if len(partials) > 0 {
		tocompile := []string{
			templateToRender,
		}
		tocompile = append(tocompile, "templates/base.layout.gohtml")
		tocompile = append(tocompile, partials...)
		t, err =
			template.New(fmt.Sprintf("%s.page.gohtml", page)).
				Funcs(functions).
				ParseFS(templateFS, tocompile...)
	} else {
		t, err =
			template.New(fmt.Sprintf("%s.page.gohtml", page)).
				Funcs(functions).
				ParseFS(templateFS, "templates/base.layout.gohtml",
					templateToRender)
	}
	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	app.templateCache[templateToRender] = t
	return t, nil
}
