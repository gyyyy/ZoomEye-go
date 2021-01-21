## ZoomEye-go

`ZoomEye` 是一款网络空间搜索引擎，用户可以使用浏览器方式 [搜索网络设备](https://www.zoomeye.org)。

`ZoomEye-go` 是一款基于 `ZoomEye API` 开发的 Golang 库，提供了 `ZoomEye` 命令行模式，同时也可以作为 `SDK` 集成到其他工具中。该库可以让技术人员更便捷地**搜索**、**筛选**、**导出** `ZoomEye` 的数据。

### 下载安装

- 在 [Releases](https://github.com/shadowsocks/go-shadowsocks2/releases) 中获得已经编译好的二进制文件

- 直接通过Github下载源代码，或运行 `go get` 进行下载安装：

       go get -u github.com/gyyyy/ZoomEye-go

   进入项目目录，执行 `make` 命令完成编译，编译好的二进制文件存放于 `bin` 目录下：

       make [all/linux/macos/win64/win32]

### 使用命令行模式

#### 基础配置

在 `ZoomEye-go` 二进制文件同目录下创建 `conf.yaml` 文件，对配置变量进行自定义设置：

```yaml
# API-Key和JWT存放路径
ZOOMEYE_CONFIG_PATH: "~/.config/zoomeye/setting"

# 搜索结果数据缓存路径
ZOOMEYE_CACHE_PATH: "~/.config/zoomeye/cache"

# 搜索和过滤结果数据保存路径
ZOOMEYE_DATA_PATH: "data"

# 本地数据超时时间，默认为5天
EXPIRED_TIME: 432000
```

若不创建或修改配置文件，`ZoomEye-go` 相关文件路径和其他参数默认值都将与 [`conf_default.yaml`](conf_default.yaml) 描述一致。

#### 初始化用户凭证

首次使用 `ZoomEye-go` 时，需要通过 `init` 命令对用户凭证进行初始化，该凭证将自动作为之后其他 `ZoomEye API` 接口调用时的身份标识。`ZoomEye-go` 实现了 `ZoomEye API` 支持的两种认证方式：

1. API-Key (推荐)

    ./ZoomEye-go init -apikey [API-KEY]

1. JWT

    ./ZoomEye-go init -username [USERNAME] -password [PASSWORD]

使用示例：

```bash
./ZoomEye-go init -apikey "XXXXXXXX-XXXX-XXXXX-XXXX-XXXXXXXXXXX"
succeed to initialize

[Resources Info]

  Role:  developer
  Quota: 10000

```

推荐使用 `API-Key` 认证方式，用户可以登录 `ZoomEye` 在 [个人信息](https://www.zoomeye.org/profile) 中获取，**注意不要将其泄露给其他人**。`JWT` 认证方式获取的凭证具有时效性，失效后需要重新初始化才能正常使用，本地存储的旧的凭证数据会被覆盖。

可以通过 `init -h` 获取帮助。

#### 查询用户资源信息

通过 `info` 命令可以查询当前用户个人信息以及数据配额，初始化成功后也会自动查询：

```bash
./ZoomEye-go info
succeed to query

[Resources Info]

  Role:  developer
  Quota: 10000

```

#### 搜索

搜索是 `ZoomEye-go` 最核心的功能，通过 `search` 命令指定搜索关键词进行使用，支持的参数说明如下：

```text
-dork [KEYWORDS]     设置搜索关键字，必填（如：-dork "port:80 nginx"）
-num [NUM]           设置显示/搜索的数据条数，默认为 20（建议设置20的倍数，因为ZoomEye一次接口查询为20条）
-type [host/web]     设置搜索资源类型，默认为 host（如：-type "web"）
-force               强制调用 ZoomEye API 查询，忽略本地数据和缓存
-count               查询该 dork 在 ZoomEye 数据库中的总量
-facet [field,...]   查询该 dork 在 ZoomEye 数据库中全量数据的分布情况，以逗号分隔（如：-facet "app,service,os"）
-stat [field,...]    统计本次搜索结果数据中指定字段的分布情况，以逗号分隔（如：-stat "app,service,os"）
-filter [field,...]  对本次搜索结果数据中指定字段进行筛选，以逗号分隔（如：-filter "app,ip,title"）
-save                保存本次搜索结果数据，若使用 filter 参数指定了筛选条件，筛选结果也会保存
```

根据搜索资源类型的不同（由参数 `-type` 确定），其他部分参数值范围存在差异，并且可能根据 `ZoomEye` 官方更新而改变：

- `-facet`与`-stat`，它们取值一样，只是统计的数据集范围不同
    - 当 `-type` 为 `host`时，可以使用 `app,device,service,os,port,country,city`
    - 当 `-type` 为 `web`时，可以使用 `webapp,component,framework,frontend,server,waf,os,country,city`
- `-filter`
    - 当 `-type` 为 `host`时，可以使用 `app,version,device,ip,port,hostname,city,city_cn,country,country_cn,asn,banner`
    - 当 `-type` 为 `web`时，可以使用 `app,headers,keywords,title,ip,site,city,city_cn,country,country_cn`

使用示例：

```bash
./ZoomEye-go search -dork "telnet" -num 1 -count
succeed to search (in 272.080753ms)

[Total Count]

  Count: 57003299


[Host Search Result]

  +-----------------------+----------------------+----------------------+------------------------------------------+----------------------+
  | Host                  | Application          | Service              | Banner                                   | Country              |
  +-----------------------+----------------------+----------------------+------------------------------------------+----------------------+
  | 159.203.16.45:10005   | Pocket CMD telnetd   | telnet               | \xff\xfb\x01\xff\xfb\x03\xff\xfc\'\xf... | Canada               |
  +-----------------------+----------------------+----------------------+------------------------------------------+----------------------+
  | Total: 1                                                                                                                              |
  +-----------------------+----------------------+----------------------+------------------------------------------+----------------------+

```

可以通过 `search -h` 获取帮助。

#### 缓存机制

`ZoomEye-go` 参考官方 `ZoomEye-python` 的设计，在命令行模式下提供了相似的缓存机制，数据默认存储在 `~/.config/zoomeye/cache` 目录，尽可能节约用户配额。搜索过的数据将默认在本地缓存 5 天，在缓存数据有效期内，重复执行同条件搜索不会消耗配额。可以设置 `-force` 参数强制调用 `ZoomEye API` 进行搜索，但结果不会覆盖当前缓存数据。

通过 `clean` 命令可以清空所有缓存数据。

### 使用SDK API

使用示例：

```go
package main

import (
	"ZoomEye-go/zoomeye"
)

func main() {
	// 初始化用户凭证
	zoom := zoomeye.New()
	jwt, _ := zoom.Login("username@zoomeye.org", "password")
	// 或使用 API-Key 进行初始化，不需要再调用 Login() 方法
	// zoom := zoomeye.NewWithKey("XXXXXXXX-XXXX-XXXXX-XXXX-XXXXXXXXXXX")

	// 查询用户资源信息
	info, _ := zoom.ResourcesInfo()

	// 搜索
	result, _ := zoom.DorkSearch("port:80 nginx", 1, "host", "app,service,os")
	// 多页搜索
	// result, _ := zoom.MultiPageSearch("wordpress country:cn", 5, "web", "webapp,server,os")

	// 对搜索结果进行统计
	stat := result.Statistics("app,service,os")

	// 对搜索结果进行筛选
	filt := result.Filter("app,ip,title")

	// 设备历史搜索（需要高级用户或VIP用户权限）
	history, _ := zoom.HistoryIP("1.2.3.4")
}
```

### TODO

- 实现交互式命令行模式
- 多页搜索优化

---

如遇到问题或Bug请提Issues，欢迎PR。
