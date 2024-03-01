package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"html/template"
	"log"
	"os"
)

const Theme = "main"

type Bot struct {
	Username     string   `json:"username"`
	ReleaseDate  string   `json:"release_date"`
	SoldDate     string   `json:"sold_date"`
	Descriptions []string `json:"descriptions"`
}

func main() {
	watch := flag.Bool("watch", false, "watch templates")
	flag.Parse()
	log.Println("watch:", *watch)

	render()

	if *watch {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("currentDir:", currentDir)

		// Рекурсивно добавляем все подпапки в вотчер
		err = watchRecursive(watcher, "./templates")
		if err != nil {
			log.Fatal("Ошибка при рекурсивном добавлении папок в вотчер:", err)
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				switch event.Op {
				case fsnotify.Write:
					log.Println("modified file:", event.Name)
					render()
				case fsnotify.Create:
					log.Println("created file:", event.Name)
					render()
				case fsnotify.Remove:
					log.Println("removed file:", event.Name)
					render()
				case fsnotify.Rename:
					log.Println("renamed file:", event.Name)
					render()
				case fsnotify.Chmod:
					log.Println("chmod file:", event.Name)
					render()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}
}

func render() {
	templatesDir := "templates"

	data := make(map[string]interface{})

	templatesDirs, err := os.ReadDir(templatesDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	templateDir := ""
	for _, tplDir := range templatesDirs {
		if tplDir.IsDir() && tplDir.Name() == Theme {
			templateDir = tplDir.Name()
		}
	}

	if templateDir == "" {
		log.Fatal("theme not found")
		return
	}

	data["Theme"] = Theme

	templateDir = templatesDir + "/" + templateDir

	// проверяем если есть папка для скомпилированных файлов
	_, err = os.Stat("./public")
	if os.IsNotExist(err) {
		err = os.Mkdir("./public", 0777)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else {
		// очищаем папку
		err = os.RemoveAll("./public")
		if err != nil {
			log.Fatal(err)
			return
		}
		err = os.Mkdir("./public", 0777)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	templateFunctions := template.FuncMap{
		"dump": func(v interface{}) string {
			return fmt.Sprintf("%+v", v)
		},
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"botList": func() []Bot {
			var bots []Bot
			// open dist file
			file, err := os.Open(templateDir + "/dist/bots.json")
			if err != nil {
				log.Fatal(err)
				return bots
			}

			// close file
			defer file.Close()

			// decode json
			err = json.NewDecoder(file).Decode(&bots)
			if err != nil {
				log.Fatal(err)
				return bots
			}
			return bots
		},
	}

	tmpl := template.New("")
	tmpl = tmpl.Funcs(templateFunctions)
	tmpl, err = tmpl.ParseGlob(templateDir + "/*.html")
	if err != nil {
		log.Fatal(err)
		return
	}

	// list parsed templates
	for _, t := range tmpl.Templates() {
		// skip empty template
		if t.Name() == "" {
			continue
		}
		log.Println("parsed template:", t.Name())
	}

	copyDir(templateDir+"/dist", "./public/dist")

	// Создаем файлы из шаблонов
	for _, tpl := range tmpl.Templates() {
		// skip empty template
		if tpl.Name() == "" {
			continue
		}
		tplName := tpl.Name()
		file, err := os.Create("./public/" + tplName)
		if err != nil {
			log.Fatal(err)
			return
		}
		_ = tpl.Execute(file, data)
	}
}
