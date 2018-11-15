# m2-helper

```sh
Usage of ./m2-helper:
  -compareTo string
    	compare to
  -cwd string
    	current working directory (default "/Users/matt/go/src/github.com/softleader/m2-helper")
  -packing string
    	determine maven packing to generate script (default "jar")
  -prefix string
    	prefix of maven deploy template
  -regex string
    	the regex to find the target file (default ".jar$")
  -repoId string
    	-DrepositoryId of 'mvn deploy:deploy-file' (default "REPO_ID")
  -suffix string
    	suffix of maven deploy template (default "-e")
  -url string
    	-Durl of 'mvn deploy:deploy-file' (default "NEXUS_URL")
```

- `compareTo` - 比較的目錄
- `cwd` - 對照的目錄, 預設當前目錄
- `regex` - 比較的檔案
- `repoId` - maven deploy 指令的 *-DrepositoryId* 參數
- `url` - maven deploy 指令的 *-Durl* 參數
- `prefix` - maven deploy template 前贅字
- `suffix` - maven deploy template 後贅字
- `packing` - 要產生 script 的 maven packing 目標

### Example

- 產生所有 jar 的 `mvn deploy:deploy-file` 指令

```sh
m2-helper -url=<nexus-url> -repoId=<server-in-settings.xml>
```

- 產生所有 packing 是 pom 的 `mvn deploy:deploy-file` 指令

```sh
m2-helper -regex=".pom$" -packing=pom -url=<nexus-url> -repoId=<server-in-settings.xml>
```

- 比較當前目錄跟指定 m2 目錄的所有 jar 檔, 並產生 script

```sh
m2-helper -compareTo=</path/to/compare/m2> -url=<nexus-url> -repoId=<server-in-settings.xml>
```
