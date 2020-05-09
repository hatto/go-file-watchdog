# Watch folder for changes

Simple script which watches for files and folders changes.
Each add/remove of file/folder will call exeternal script - in this case it will be a wp-cli script ```wp loop media [add|remove] /changed/filed/path```

use:
```
fileChangeListener /path/to/my/folder
```


build for linux
```
env GOOS=linux GOARCH=amd64 GOARM=7 go build fileChangeListener.go
```

buil for os x (local)
```
go build fileChangeListener.go
```
