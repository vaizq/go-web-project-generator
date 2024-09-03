package main

import (
	"bufio"
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed templates/*
var templates embed.FS

type config struct {
	Project_name string
	DSN          string
}

type application struct {
	cfg *config
}

func hello() {
	fmt.Println("### GO webserver-project generator ###")
}

func goodBye() {
	fmt.Println("Project was initialized successfully!\nRun your site with: go run ./cmd/web")
}

func main() {
	hello()

	app, err := newApplication()
	if err != nil {
		log.Fatal("Failed to create application")
	}

	app.run()
	goodBye()
}

func newConfig() (*config, error) {
	conf := &config{
		Project_name: askInput("Project name? "),
		DSN:          askInput("Database DSN? "),
	}
	return conf, nil
}

func newApplication() (*application, error) {
	config, err := newConfig()
	if err != nil {
		log.Fatal("Failed to get config")
	}

	return &application{cfg: config}, nil
}

func askInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func executeCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Failed to execute command: ", err.Error())
		return err
	}

	fmt.Printf("%s\n", output)

	return nil
}

func (app *application) makeFile(filename string) {
	templateName := filepath.Base(filename)
	templateName = strings.Split(templateName, ".")[0] + ".tmpl"
	loaded := app.loadTemplate(templateName)

	err := os.MkdirAll(filepath.Dir(filename), 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	err = os.WriteFile(filename, []byte(loaded), 0660)
	if err != nil {
		log.Fatal(err)
	}
}

func (app *application) makeHiddenFile(filename string) {
	tname := filepath.Base(filename)

	if tname[0] != '.' {
		log.Fatalf("%s is not an hidden filepath", filename)
	}

	if strings.Count(tname, ".") > 1 {
		tname = "." + strings.Split(tname, ".")[1] + ".tmpl"
	} else {
		tname = tname + ".tmpl"
	}

	loaded := app.loadTemplate(tname)

	err := os.MkdirAll(filepath.Dir(filename), 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	err = os.WriteFile(filename, []byte(loaded), 0660)
	if err != nil {
		log.Fatal(err)
	}
}

func makeDir(path string) {
	err := os.MkdirAll(path, 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
}

func (app *application) loadTemplate(filename string) string {
	ts, err := template.ParseFS(templates, fmt.Sprintf("templates/%s", filename))
	if err != nil {
		log.Fatal(err)
	}

	var b strings.Builder

	err = ts.Execute(&b, app.cfg)
	if err != nil {
		log.Fatal(err)
	}

	return b.String()
}

func (app *application) run() {
	app.makeFile("cmd/web/main.go")
	app.makeFile("cmd/web/routes.go")
	app.makeFile("cmd/web/handlers.go")
	app.makeFile("cmd/web/middleware.go")
	app.makeFile("cmd/web/templates.go")
	app.makeFile("cmd/web/helpers.go")
	app.makeFile("internal/model/models.go")
	app.makeFile("ui/html/base.html")
	app.makeFile("ui/html/partials/nav.html")
	app.makeFile("ui/html/pages/index.html")
	app.makeHiddenFile(".env")

	makeDir("db/migrations")
	makeDir("static/css")
	makeDir("static/js")
	makeDir("static/img")

	executeCommand(fmt.Sprintf("go mod init %s && go mod tidy", app.cfg.Project_name))
}
