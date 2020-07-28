## logs

```go
package main

import (
    "context"

    "skylar-lib/logs/configor"
    "skylar-lib/logs/store/logs"
)

type Config struct {
    Logs    logs.Config
}

var (
    conf Config
)

func init() {
    if err := configor.Load("./configs/conf.toml", &conf); err != nil {
        panic(err)
    }

    logs.Init(conf.Logs)
}

func main() {
    logs.Info("hello world")
    logs.Infof("hello %s", "world")
}
```

```toml
#conf.toml - file
[logs]
    writer = "file"
    level  = "info"
    [logs.file_config]
        path        = "/tmp/logs/service.log"
        compress    = false
        max_size    = 1024            # 1G
        max_age     = 30              # 30 days
        max_backups = 1               # 1 copy

#conf.toml - console
[logs]
    writer = "console"
    level  = "info"

#conf.toml - file & console
[logs]
    writer = "file,console"
    level  = "info"
    [logs.file_config]
        path        = "/tmp/logs/service.log"
        compress    = false
        max_size    = 1024            # 1G
        max_age     = 30              # 30 days
        max_backups = 1               # 1 copy
```
