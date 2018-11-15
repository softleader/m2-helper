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
	mvnDeploy = `mvn deploy:deploy-file -DgroupId={{ .GetGroupId }} -DartifactId={{ .GetArtifactId }} -Dversion={{ .GetVersion }} -Dpackaging={{ .GetPackaging }} -Dfile={{ .File }} -DrepositoryId={{ .RepositoryId }} -Durl={{ .Url }} -e`
)

var notFounds []string
var sizeWrongs []string
var scripts []string
var pomErrors []string
var root string
var compareTo string
var target *regexp.Regexp

func main() {
	root, _ = os.Getwd()
	flag.StringVar(&compareTo, "compareTo", "", "compare to")
	flag.StringVar(&root, "cwd", root, "current working directory")
	regex := flag.String("regex", ".jar$", "the regex to find the target file")
	flag.Parse()
	target = regexp.MustCompile(*regex)

	walkDir(root, containsTargetFile, compareFile)

	if len(notFounds) > 0 {
		fmt.Printf("\n檔案不存在\n")
		for _, notfound := range notFounds {
			fmt.Println(notfound)
		}
	}

	if len(sizeWrongs) > 0 {
		fmt.Printf("\n檔案大小 size 不合\n")
		for _, sizewrong := range sizeWrongs {
			fmt.Println(sizewrong)
		}
	}

	if len(pomErrors) > 0 {
		fmt.Printf("\nLoad POM Error\n")
		for _, pomerror := range pomErrors {
			fmt.Println(pomerror)
		}
	}

	fmt.Printf("\n")
	scripts = distinct(scripts)
	for _, script := range scripts {
		fmt.Println(script)
	}
}

func walkDir(dirpath string, stop func(path string) bool, callback func(path string)) {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			p := path.Join(dirpath, file.Name())
			if stop(p) {
				callback(p)
			} else {
				walkDir(p, stop, callback)
			}
		}
	}
}

func compareFile(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			expected := filepath.Join(path, f.Name())
			if target.MatchString(f.Name()) {
				if compareTo != "" {
					actual := filepath.Join(strings.Replace(path, root, compareTo, 1), f.Name())
					compare(expected, actual)
				}
				generateScript(path)
			}
		}
	}
}

func generateScript(path string) {
	pomFile := searchPomFile(path)
	pom := loadPom(pomFile)

	if pom.isPackingPom() {
		pom.File = strings.Replace(pomFile, root, ".", 1)
	} else {
		pom.File = searchJarFile(path)
		if pom.File == "" { // pom 宣告是 packing jar, 但目錄下又找不到 jar??
			return
		}
		pom.File = strings.Replace(pom.File, root, ".", 1)
	}

	pom.RepositoryId = "REPO_ID"
	pom.Url = "NEXUS_URL"

	var buf bytes.Buffer
	t := template.Must(template.New("").Parse(mvnDeploy))
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

		decoder := xml.NewDecoder(strings.NewReader(string(bytes)))
		decoder.CharsetReader = charset.NewReaderLabel
		err = decoder.Decode(&pom)

		if err != nil {
			pomErrors = append(pomErrors, err.Error()+": "+path)
		}
	}
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

func containsTargetFile(path string) bool {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if !f.IsDir() && target.MatchString(f.Name()) {
			return true
		}
	}
	return false
}

type Pom struct {
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

func (p *Pom) isPackingPom() bool {
	return p.GetPackaging() == "pom"
}

func (p *Pom) isJar() bool {
	return p.GetPackaging() == "jar"
}