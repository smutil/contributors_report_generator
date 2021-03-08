package main

import (
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Global      Global
	GitRepoList []Git `yaml:"git"`
}

type Global struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Since    string `yaml:"since"`
}

type Git struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type CommitAuthor struct {
	Name        string
	Email       string
	CommitCount int
}

var Version = "develop"

func main() {
	var configPath string
	var version bool
	flag.BoolVar(&version, "version", false, "returns the fileuploader version")
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.Parse()
	if version {
		fmt.Println(Version)
		return
	}
	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		log.Fatal(err)
	}
	config := &Config{}
	err := ReadYML(configPath, &config)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	generateMetrics(*config)
}

func generateMetrics(config Config) {
	var scm_repo, scm_usr, scm_pwd, since string
	scm_usr = config.Global.Username
	scm_pwd = config.Global.Password
	since = config.Global.Since
	var m = make(map[string]CommitAuthor)

	for _, gitrepo := range config.GitRepoList {
		scm_repo = gitrepo.URL
		log.Println("retreiving commitors for : " + scm_repo)
		r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL: scm_repo,
			Auth: &http.BasicAuth{
				Username: scm_usr,
				Password: scm_pwd,
			},
		})
		if err != nil {
			fmt.Sprintf("some error occureded for "+scm_repo, err)
		} else {
			// ... retrieving all commits
			ref, _ := r.Head()
			if since == "" {
				since = "02.01.2006"
			}
			since, err := time.Parse("02.01.2006", since)
			if err != nil {
				fmt.Println(err)
				return
			}
			cIter, _ := r.Log(&git.LogOptions{From: ref.Hash(), Since: &since})

			cIter.ForEach(func(c *object.Commit) error {
				var i = m[c.Author.Email].CommitCount
				if i == 0 {
					m[c.Author.Email] = CommitAuthor{c.Author.Name, c.Author.Email, 1}
				} else {
					m[c.Author.Email] = CommitAuthor{c.Author.Name, c.Author.Email, i + 1}
				}
				return nil
			})
		}
	}
	fmt.Println(m)
	log.Println("INFO : creating report")
	f := excelize.NewFile()

	// define the border style
	border := []excelize.Border{
		{Type: "top", Style: 2, Color: "cccccc"},
		{Type: "left", Style: 2, Color: "cccccc"},
		{Type: "right", Style: 2, Color: "cccccc"},
		{Type: "bottom", Style: 2, Color: "cccccc"},
	}
	// define the style of the header row
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{
			Type: "pattern", Color: []string{"dae9f3"}, Pattern: 1},
		Border: border},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// define the style of cells
	cellsStyle, err := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Color: "333333"},
		Border: border})

	if err != nil {
		fmt.Println(err)
		return
	}
	var i int = 1
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Email")
	f.SetCellValue("Sheet1", "C1", "Commit Count")
	f.SetCellStyle("Sheet1", "A1", "C1", headerStyle)
	f.SetColWidth("Sheet1", "A", "B", 25)
	f.SetColWidth("Sheet1", "C", "C", 15)
	for k, v := range m {
		i = i + 1
		var row = strconv.Itoa(i)
		f.SetCellValue("Sheet1", "A"+row, v.Name)
		f.SetCellValue("Sheet1", "B"+row, k)
		f.SetCellValue("Sheet1", "C"+row, v.CommitCount)
		f.SetCellStyle("Sheet1", "A"+row, "C"+row, cellsStyle)
	}
	if err := f.SaveAs("contributors_report.xlsx"); err != nil {
		fmt.Println(err)
	}
}

func readfile(scm_repo string, scm_usr string, scm_pwd string) {
	fs := memfs.New()
	storer := memory.NewStorage()
	_, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: scm_repo,
		Auth: &http.BasicAuth{
			Username: scm_usr,
			Password: scm_pwd,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	changelog, err := fs.Open("README.md")
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(os.Stdout, changelog)
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func ExitIfError(err error) {
	if err == nil {
		return
	}
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

func ReadYML(configPath string, configPointer interface{}) error {
	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Init new YAML decoder
	d := yaml.NewDecoder(file)
	if err := d.Decode(configPointer); err != nil {
		return err
	}

	return nil
}

func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}
