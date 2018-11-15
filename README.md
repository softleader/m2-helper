# m2-helper

```sh
Usage of ./m2-helper:
  -compareTo string
    	compare to
  -cwd string
    	current working directory
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