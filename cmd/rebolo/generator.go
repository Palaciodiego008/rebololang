package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed templates
var templates embed.FS

type Generator struct {
	templates *template.Template
}

type AppData struct {
	Name      string
	Module    string
	Framework string
	Title     string
}

type ResourceData struct {
	Name      string
	VarName   string
	TableName string
	ViewPath  string
	RoutePath string
	Fields    []Field
	Timestamp string
}

type Field struct {
	Name     string
	DBName   string
	FormName string
	GoType   string
	SQLType  string
	HTMLType string
}

func NewGenerator() *Generator {
	// Parse all template files recursively
	tmpl := template.New("").Funcs(template.FuncMap{
		"title": func(s string) string { return cases.Title(language.English).String(s) },
		"lower": strings.ToLower,
	})

	// Parse templates manually to handle nested directories
	tmpl = template.Must(tmpl.ParseFS(templates,
		"templates/app/main.go.tmpl",
		"templates/app/package.json.tmpl",
		"templates/app/src/index.js.tmpl",
		"templates/app/views/layouts/application.html.tmpl",
		"templates/app/views/home/index.html.tmpl",
		"templates/config/config.yml.tmpl",
		"templates/resource/model.go.tmpl",
		"templates/resource/controller.go.tmpl",
		"templates/resource/migration.sql.tmpl",
	))

	return &Generator{templates: tmpl}
}

func (g *Generator) GenerateApp(name string) error {
	data := AppData{
		Name:      name,
		Module:    fmt.Sprintf("github.com/Palaciodiego008/%s", name),
		Framework: "ReboloLang",
		Title:     fmt.Sprintf("Welcome to %s", name),
	}

	// Create directory structure
	dirs := []string{
		name,
		filepath.Join(name, "controllers"),
		filepath.Join(name, "models"),
		filepath.Join(name, "views", "home"),
		filepath.Join(name, "views", "layouts"),
		filepath.Join(name, "public"),
		filepath.Join(name, "src"),
		filepath.Join(name, "db", "migrations"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Generate files from templates
	files := map[string]string{
		filepath.Join(name, "main.go"):                              "app/main.go.tmpl",
		filepath.Join(name, "package.json"):                         "app/package.json.tmpl",
		filepath.Join(name, "config.yml"):                           "config/config.yml.tmpl",
		filepath.Join(name, "src", "index.js"):                      "app/src/index.js.tmpl",
		filepath.Join(name, "views", "layouts", "application.html"): "app/views/layouts/application.html.tmpl",
		filepath.Join(name, "views", "home", "index.html"):          "app/views/home/index.html.tmpl",
	}

	for filePath, tmplName := range files {
		if err := g.renderTemplate(tmplName, filePath, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filePath, err)
		}
	}

	fmt.Printf("✅ Generated app: %s\n", name)
	return nil
}

func (g *Generator) GenerateResource(name string, fieldArgs []string) error {
	fields := g.parseFields(fieldArgs)

	data := ResourceData{
		Name:      cases.Title(language.English).String(name),
		VarName:   strings.ToLower(name),
		TableName: g.pluralize(strings.ToLower(name)),
		ViewPath:  g.pluralize(strings.ToLower(name)),
		RoutePath: g.pluralize(strings.ToLower(name)),
		Fields:    fields,
		Timestamp: time.Now().Format("20060102150405"),
	}

	// Create directories
	os.MkdirAll("models", 0755)
	os.MkdirAll("controllers", 0755)
	os.MkdirAll("db/migrations", 0755)
	os.MkdirAll(filepath.Join("views", data.ViewPath), 0755)

	// Generate files
	files := map[string]string{
		filepath.Join("models", data.VarName+".go"):                                        "resource/model.go.tmpl",
		filepath.Join("controllers", data.VarName+"_controller.go"):                        "resource/controller.go.tmpl",
		filepath.Join("db", "migrations", data.Timestamp+"_create_"+data.TableName+".sql"): "resource/migration.sql.tmpl",
	}

	for filePath, tmplName := range files {
		if err := g.renderTemplate(tmplName, filePath, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filePath, err)
		}
	}

	// Generate views
	if err := g.generateViews(data); err != nil {
		return err
	}

	fmt.Printf("✅ Generated resource: %s\n", name)
	fmt.Printf("   - Model: models/%s.go\n", data.VarName)
	fmt.Printf("   - Controller: controllers/%s_controller.go\n", data.VarName)
	fmt.Printf("   - Migration: db/migrations/%s_create_%s.sql\n", data.Timestamp, data.TableName)
	fmt.Printf("   - Views: views/%s/\n", data.ViewPath)

	return nil
}

func (g *Generator) renderTemplate(tmplName, filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Extract just the filename from the template path
	parts := strings.Split(tmplName, "/")
	templateName := parts[len(parts)-1]

	return g.templates.ExecuteTemplate(file, templateName, data)
}

func (g *Generator) parseFields(fieldArgs []string) []Field {
	var fields []Field

	for _, arg := range fieldArgs {
		parts := strings.Split(arg, ":")
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		fieldType := parts[1]

		field := Field{
			Name:     cases.Title(language.English).String(name),
			DBName:   strings.ToLower(name),
			FormName: strings.ToLower(name),
			GoType:   g.mapToGoType(fieldType),
			SQLType:  g.mapToSQLType(fieldType),
			HTMLType: g.mapToHTMLType(fieldType),
		}

		fields = append(fields, field)
	}

	return fields
}

func (g *Generator) mapToGoType(dbType string) string {
	switch dbType {
	case "string", "text":
		return "string"
	case "int", "integer":
		return "int64"
	case "bool", "boolean":
		return "bool"
	case "float":
		return "float64"
	case "time", "datetime":
		return "time.Time"
	default:
		return "string"
	}
}

func (g *Generator) mapToSQLType(goType string) string {
	switch goType {
	case "string":
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "int", "integer":
		return "BIGINT"
	case "bool", "boolean":
		return "BOOLEAN"
	case "float":
		return "DECIMAL"
	case "time", "datetime":
		return "TIMESTAMP"
	default:
		return "VARCHAR(255)"
	}
}

func (g *Generator) mapToHTMLType(goType string) string {
	switch goType {
	case "string":
		return "text"
	case "text":
		return "textarea"
	case "int", "integer":
		return "number"
	case "bool", "boolean":
		return "checkbox"
	case "float":
		return "number"
	case "time", "datetime":
		return "datetime-local"
	default:
		return "text"
	}
}

func (g *Generator) pluralize(word string) string {
	// Simple pluralization
	if strings.HasSuffix(word, "s") {
		return word + "es"
	}
	return word + "s"
}

func (g *Generator) generateViews(data ResourceData) error {
	views := map[string]string{
		"index.html": g.generateIndexView(data),
		"show.html":  g.generateShowView(data),
		"new.html":   g.generateNewView(data),
		"edit.html":  g.generateEditView(data),
	}

	for filename, content := range views {
		filePath := filepath.Join("views", data.ViewPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateIndexView(data ResourceData) string {
	return fmt.Sprintf(`<h1>%s</h1>
<a href="/%s/new" class="btn">New %s</a>

<div style="margin-top: 2rem;">
    {{range .%s}}
    <div style="border: 1px solid rgba(255,255,255,0.2); padding: 1rem; margin: 1rem 0; border-radius: 8px; background: rgba(255,255,255,0.1);">
        <h3><a href="/%s/{{.ID}}" style="color: white;">{{.%s}}</a></h3>
        <div style="margin-top: 10px;">
            <a href="/%s/{{.ID}}/edit" class="btn" style="background: #2196F3;">Edit</a>
            <form method="POST" action="/%s/{{.ID}}" style="display: inline; margin-left: 10px;">
                <input type="hidden" name="_method" value="DELETE">
                <button type="submit" onclick="return confirm('Are you sure?')" class="btn" style="background: #f44336;">Delete</button>
            </form>
        </div>
    </div>
    {{end}}
</div>`,
		cases.Title(language.English).String(data.TableName), data.RoutePath, data.Name,
		cases.Title(language.English).String(data.TableName), data.RoutePath, g.getFirstStringField(data.Fields),
		data.RoutePath, data.RoutePath)
}

func (g *Generator) generateShowView(data ResourceData) string {
	fieldsHTML := ""
	for _, field := range data.Fields {
		fieldsHTML += fmt.Sprintf(`    <p><strong>%s:</strong> {{.%s}}</p>
`, field.Name, field.Name)
	}

	return fmt.Sprintf(`<h1>%s Details</h1>
<div style="background: rgba(255,255,255,0.1); padding: 2rem; border-radius: 8px; margin: 1rem 0; text-align: left;">
%s</div>
<div>
    <a href="/%s/{{.ID}}/edit" class="btn" style="background: #2196F3;">Edit</a>
    <a href="/%s" class="btn" style="background: #666; margin-left: 10px;">Back to List</a>
</div>`,
		data.Name, fieldsHTML, data.RoutePath, data.RoutePath)
}

func (g *Generator) generateNewView(data ResourceData) string {
	return fmt.Sprintf(`<h1>New %s</h1>
<form method="POST" action="/%s" style="max-width: 500px; text-align: left;">
%s
    <div style="margin-top: 1rem;">
        <button type="submit" class="btn">Create %s</button>
        <a href="/%s" class="btn" style="background: #666; margin-left: 10px;">Cancel</a>
    </div>
</form>`,
		data.Name, data.RoutePath, g.generateFormFields(data.Fields), data.Name, data.RoutePath)
}

func (g *Generator) generateEditView(data ResourceData) string {
	return fmt.Sprintf(`<h1>Edit %s</h1>
<form method="POST" action="/%s/{{.ID}}" style="max-width: 500px; text-align: left;">
    <input type="hidden" name="_method" value="PUT">
%s
    <div style="margin-top: 1rem;">
        <button type="submit" class="btn" style="background: #2196F3;">Update %s</button>
        <a href="/%s/{{.ID}}" class="btn" style="background: #666; margin-left: 10px;">Cancel</a>
    </div>
</form>`,
		data.Name, data.RoutePath, g.generateFormFields(data.Fields), data.Name, data.RoutePath)
}

func (g *Generator) generateFormFields(fields []Field) string {
	html := ""
	for _, field := range fields {
		if field.HTMLType == "textarea" {
			html += fmt.Sprintf(`    <div class="form-group">
        <label><strong>%s:</strong></label>
        <textarea name="%s" rows="4"></textarea>
    </div>
`, field.Name, field.FormName)
		} else if field.HTMLType == "checkbox" {
			html += fmt.Sprintf(`    <div class="form-group">
        <label><input type="checkbox" name="%s" value="true"> <strong>%s</strong></label>
    </div>
`, field.FormName, field.Name)
		} else {
			html += fmt.Sprintf(`    <div class="form-group">
        <label><strong>%s:</strong></label>
        <input type="%s" name="%s">
    </div>
`, field.Name, field.HTMLType, field.FormName)
		}
	}
	return html
}

func (g *Generator) getFirstStringField(fields []Field) string {
	for _, field := range fields {
		if field.GoType == "string" {
			return field.Name
		}
	}
	return "ID"
}
