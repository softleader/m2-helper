package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"golang.org/x/net/html/charset"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	mvnDeployTemplate = `{{ .Prefix }} mvn deploy:deploy-file -DgroupId={{ .GetGroupId }} -DartifactId={{ .GetArtifactId }} -Dversion={{ .GetVersion }} -Dpackaging={{ .GetPackaging }} -Dfile={{ .File }} -DrepositoryId={{ .RepositoryId }} -Durl={{ .Url }} {{ .Suffix }}`
)

var notFounds []string
var sizeWrongs []string
var scripts []string
var pomErrors []string
var root string
var compareTo string
var target *regexp.Regexp
var url string
var prefix string
var suffix string
var repoId string
var packing string

func main() {
	root, _ = os.Getwd()
	flag.StringVar(&compareTo, "compareTo", "", "compare to")
	flag.StringVar(&root, "cwd", root, "current working directory")
	flag.StringVar(&url, "url", "NEXUS_URL", "-Durl of 'mvn deploy:deploy-file'")
	flag.StringVar(&repoId, "repoId", "REPO_ID", "-DrepositoryId of 'mvn deploy:deploy-file'")
	flag.StringVar(&prefix, "prefix", "", "prefix of maven deploy template")
	flag.StringVar(&suffix, "suffix", "-e", "suffix of maven deploy template")
	flag.StringVar(&packing, "packing", "jar", "determine maven packing to generate script")
	regex := flag.String("regex", ".jar$", "the regex to find the target file")

	flag.Parse()
	target = regexp.MustCompile(*regex)

	walkDir(root, func(path string) (stop bool) {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if !f.IsDir() && target.MatchString(f.Name()) {
				stop = true
				expected := filepath.Join(path, f.Name())
				if compareTo != "" {
					actual := filepath.Join(strings.Replace(path, root, compareTo, 1), f.Name())
					compare(expected, actual)
				} else {
					if !strings.HasSuffix(f.Name(), ".pom") {
						expected = searchPomFile(path)
					}
					pom := loadPom(expected)
					generateScript(pom)
				}

			}
		}
		return
	})

	if len(notFounds) > 0 {
		fmt.Printf("\n檔案不存在\n")
		for _, notFound := range notFounds {
			fmt.Println(notFound)
		}
	}

	if len(sizeWrongs) > 0 {
		fmt.Printf("\n檔案大小 size 不合\n")
		for _, sizeWrong := range sizeWrongs {
			fmt.Println(sizeWrong)
		}
	}

	if len(pomErrors) > 0 {
		fmt.Printf("\nLoad POM Error\n")
		for _, pomError := range pomErrors {
			fmt.Println(pomError)
		}
	}

	fmt.Printf("\n")
	scripts = distinct(scripts)
	for _, script := range scripts {
		fmt.Println(script)
	}
}

func walkDir(dirpath string, stop func(path string) bool) {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			if p := path.Join(dirpath, file.Name()); !stop(p) {
				walkDir(p, stop)
			}
		}
	}
}

func generateScript(pom Pom) {
	if pom.Packaging != packing { // not the generate target
		return
	}
	if pom.Packaging == "pom" {
		pom.File = strings.Replace(pom.Path, root, ".", 1)
	} else if pom.Packaging == "jar" {
		pom.File = searchJarFile(filepath.Dir(pom.Path))
		if pom.File == "" { // pom 宣告是 packing jar, 但目錄下又找不到 jar??
			return
		}
		pom.File = strings.Replace(pom.File, root, ".", 1)
	} else {
		panic(fmt.Sprintf("Unsupported maven packing: %s", pom.Packaging))
	}
	var buf bytes.Buffer
	t := template.Must(template.New("").Parse(mvnDeployTemplate))
	err := t.Execute(&buf, pom)
	if err != nil {
		panic(err)
	}
	scripts = append(scripts, buf.String())
}

func distinct(s []string) []string {
	unique := make(map[string]bool, len(s))
	us := make([]string, len(unique))
	for _, elem := range s {
		if len(elem) != 0 {
			if !unique[elem] {
				us = append(us, elem)
				unique[elem] = true
			}
		}
	}
	return us
}

func searchPomFile(path string) string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pom") {
			return filepath.Join(path, f.Name())
		}
	}
	panic(fmt.Errorf("POM not found under %s\n", path))
}

func searchJarFile(path string) string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".jar") {
			return filepath.Join(path, f.Name())
		}
	}
	return ""
}

func loadPom(path string) (pom Pom) {
	xmlFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer xmlFile.Close()
	bytes, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		panic(err)
	}
	err = xml.Unmarshal(bytes, &pom)
	if err != nil {
		// xml unmarshal 只支援 UTF8, 如 ISO-8859-1 的就要用 decoder 轉換
		decoder := xml.NewDecoder(strings.NewReader(string(bytes)))
		decoder.CharsetReader = charset.NewReaderLabel
		err = decoder.Decode(&pom)

		if err != nil {
			pomErrors = append(pomErrors, err.Error()+": "+path)
		}
	}

	pom.Path = path
	pom.RepositoryId = repoId
	pom.Url = url
	pom.Prefix = prefix
	pom.Suffix = suffix

	return
}

func compare(expectedPath string, actualPath string) {
	expected, err := os.Stat(expectedPath)
	if err != nil {
		panic(err)
	}

	actual, err := os.Stat(actualPath)
	if err != nil {
		notFounds = append(notFounds, strings.Replace(actualPath, compareTo, ".", 1))
	} else {
		if expected.Size() != actual.Size() {
			sizeWrongs = append(sizeWrongs, fmt.Sprintf("%s (預期: %v, 實際: %v)", strings.Replace(actualPath, compareTo, ".", 1), expected.Size(), actual.Size()))
		}
	}
}

type Pom struct {
	Path   string // pom 檔案的路徑
	Parent struct {
		GroupId    string `xml:"groupId"`
		ArtifactId string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"parent"`
	GroupId      string `xml:"groupId"`
	ArtifactId   string `xml:"artifactId"`
	Version      string `xml:"version"`
	Packaging    string `xml:"packaging"`
	File         string
	RepositoryId string
	Url          string
	Prefix       string
	Suffix       string
}

func (p Pom) GetGroupId() (s string) {
	s = p.GroupId
	if s == "" {
		s = p.Parent.GroupId
	}
	if s == "" {
		panic(fmt.Sprintf("Can not find groupId of %v\n", p))
	}
	return
}

func (p Pom) GetArtifactId() (s string) {
	s = p.ArtifactId
	if s == "" {
		s = p.Parent.ArtifactId
	}
	if s == "" {
		panic(fmt.Sprintf("Can not find artifactId of %v\n", p))
	}
	return
}

func (p Pom) GetVersion() (s string) {
	s = p.Version
	if s == "" {
		s = p.Parent.Version
	}
	if s == "" {
		panic(fmt.Sprintf("Can not find version of %v\n", p))
	}
	return
}

func (p Pom) GetPackaging() string {
	if p.Packaging == "" {
		return "jar"
	}
	return p.Packaging
}
