package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
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
	templates   *template.Template
	typeMapping *FieldTypeMapping
}

type AppData struct {
	Name      string
	Module    string
	Framework string
	Title     string
}

type ResourceData struct {
	Name       string
	VarName    string
	Module     string
	TableName  string
	ViewPath   string
	RoutePath  string
	Fields     []Field
	FirstField string
	Timestamp  string
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
		"templates/app/src/styles.css.tmpl",
		"templates/app/views/layouts/application.html.tmpl",
		"templates/app/views/home/index.html.tmpl",
		"templates/config/config.yml.tmpl",
		"templates/resource/model.go.tmpl",
		"templates/resource/controller.go.tmpl",
		"templates/resource/migration.sql.tmpl",
	))

	return &Generator{
		templates:   tmpl,
		typeMapping: DefaultFieldTypeMapping(),
	}
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
		filepath.Join(name, "src", "styles.css"):                    "app/src/styles.css.tmpl",
		filepath.Join(name, "views", "layouts", "application.html"): "app/views/layouts/application.html.tmpl",
		filepath.Join(name, "views", "home", "index.html"):          "app/views/home/index.html.tmpl",
	}

	for filePath, tmplName := range files {
		if err := g.renderTemplate(tmplName, filePath, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filePath, err)
		}
	}

	// Initialize go.mod (like Buffalo does)
	fmt.Printf("📦 Initializing Go module...\n")
	cmd := exec.Command("go", "mod", "init", name)
	cmd.Dir = name
	if err := cmd.Run(); err != nil {
		// If go.mod already exists or error, continue (user might have created it manually)
		fmt.Printf("⚠️  Note: go mod init skipped (module may already exist)\n")
	}

	fmt.Printf("✅ Generated app: %s\n", name)
	fmt.Printf("💡 Next steps:\n")
	fmt.Printf("   cd %s\n", name)
	fmt.Printf("   go mod tidy\n")
	fmt.Printf("   rebolo dev\n")
	return nil
}

func (g *Generator) GenerateResource(name string, fieldArgs []string) error {
	fields := g.parseFields(fieldArgs)

	// Get module name from go.mod
	moduleName := g.getModuleName()

	data := ResourceData{
		Name:       cases.Title(language.English).String(name),
		VarName:    strings.ToLower(name),
		Module:     moduleName,
		TableName:  g.pluralize(strings.ToLower(name)),
		ViewPath:   g.pluralize(strings.ToLower(name)),
		RoutePath:  g.pluralize(strings.ToLower(name)),
		Fields:     fields,
		FirstField: g.getFirstStringField(fields),
		Timestamp:  time.Now().Format("20060102150405"),
	}

	// Create directories
	os.MkdirAll("models", 0755)
	os.MkdirAll("controllers", 0755)
	os.MkdirAll("db/migrations", 0755)
	os.MkdirAll(filepath.Join("views", data.ViewPath), 0755)

	// Generate files (models, controllers, migrations)
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

	// Generate views using separate template instances to avoid name conflicts
	if err := g.generateResourceViews(data); err != nil {
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
	if goType, ok := g.typeMapping.GoTypes[dbType]; ok {
		return goType
	}
	return "string" // default fallback
}

func (g *Generator) mapToSQLType(goType string) string {
	if sqlType, ok := g.typeMapping.SQLTypes[goType]; ok {
		return sqlType
	}
	return "VARCHAR(255)" // default fallback
}

func (g *Generator) mapToHTMLType(goType string) string {
	if htmlType, ok := g.typeMapping.HTMLTypes[goType]; ok {
		return htmlType
	}
	return "text" // default fallback
}

func (g *Generator) pluralize(word string) string {
	// Enhanced pluralization rules
	switch {
	case strings.HasSuffix(word, "s"), strings.HasSuffix(word, "x"), strings.HasSuffix(word, "z"):
		return word + "es"
	case strings.HasSuffix(word, "ch"), strings.HasSuffix(word, "sh"):
		return word + "es"
	case strings.HasSuffix(word, "y"):
		// Check if preceded by consonant
		if len(word) > 1 && !isVowel(rune(word[len(word)-2])) {
			return word[:len(word)-1] + "ies"
		}
		return word + "s"
	case strings.HasSuffix(word, "f"):
		return word[:len(word)-1] + "ves"
	case strings.HasSuffix(word, "fe"):
		return word[:len(word)-2] + "ves"
	default:
		return word + "s"
	}
}

func isVowel(r rune) bool {
	return strings.ContainsRune("aeiouAEIOU", r)
}

func (g *Generator) generateResourceViews(data ResourceData) error {
	viewTemplates := map[string]string{
		"index.html": "templates/resource/index.html.tmpl",
		"show.html":  "templates/resource/show.html.tmpl",
		"new.html":   "templates/resource/new.html.tmpl",
		"edit.html":  "templates/resource/edit.html.tmpl",
	}

	for filename, tmplPath := range viewTemplates {
		// Read template content
		tmplContent, err := templates.ReadFile(tmplPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", tmplPath, err)
		}

		// Create new template instance for each view
		tmpl, err := template.New(filename).Funcs(template.FuncMap{
			"title": func(s string) string { return cases.Title(language.English).String(s) },
			"lower": strings.ToLower,
		}).Parse(string(tmplContent))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", filename, err)
		}

		// Generate the view file
		filePath := filepath.Join("views", data.ViewPath, filename)
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", filePath, err)
		}
		defer file.Close()

		if err := tmpl.Execute(file, data); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", filename, err)
		}
	}

	return nil
}

func (g *Generator) getFirstStringField(fields []Field) string {
	for _, field := range fields {
		if field.GoType == "string" {
			return field.Name
		}
	}
	return "ID"
}

func (g *Generator) getModuleName() string {
	// Read go.mod to get module name
	data, err := os.ReadFile("go.mod")
	if err != nil {
		// Fallback to directory name if go.mod doesn't exist
		dir, err := os.Getwd()
		if err != nil {
			return "app"
		}
		return filepath.Base(dir)
	}

	// Parse module name from go.mod (first line: "module <name>")
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}

	// Fallback
	return "app"
}
