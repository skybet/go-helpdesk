<p align="center"><a href="#readme"><img src="https://gh.kaos.st/go-slack.svg"/></a></p>

<p align="center"><a href="#installing">Installing</a> • <a href="#examples">Examples</a> • <a href="#build-status">Build Status</a> • <a href="#license">License</a></p>

<p align="center">
  <a href="https://godoc.org/pkg.re/essentialkaos/slack.v3"><img src="https://godoc.org/github.com/essentialkaos/slack?status.svg"></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/slack"><img src="https://goreportcard.com/badge/github.com/essentialkaos/slack"></a>
  <a href="https://codebeat.co/projects/github-com-essentialkaos-slack-master"><img src="https://codebeat.co/badges/ebb05788-d2f3-4299-80cd-0dcb84e55e4d"></a>
  <a href="https://travis-ci.org/essentialkaos/slack"><img src="https://travis-ci.org/essentialkaos/slack.svg"></a>
  <a href="#license"><img src="https://gh.kaos.st/bsd.svg"></a>
</p>

This library supports most if not all of the `api.slack.com` REST calls, as well as the Real-Time Messaging protocol over websocket, in a fully managed way.

* [Installing](#installing)
* [Examples](#examples)
* [Build Status](#build-status)
* [License](#license)

### Installing

Before the initial install allows git to use redirects for [pkg.re](https://github.com/essentialkaos/pkgre) service (_reason why you should do this described [here](https://github.com/essentialkaos/pkgre#git-support)_):

```
git config --global http.https://pkg.re.followRedirects true
```

Make sure you have a working Go 1.7+ workspace ([instructions](https://golang.org/doc/install)), then:

```
go get pkg.re/essentialkaos/slack.v3
```

For update to latest stable release, do:

```
go get -u pkg.re/essentialkaos/slack.v3
```

### Examples

#### Getting all groups

```golang
import (
  "fmt"

  "pkg.re/essentialkaos/slack.v3"
)

func main() {
  api := slack.New("YOUR_TOKEN_HERE")
  
  // If you set debugging, it will log all requests to the console
  // Useful when encountering issues
  // api.SetDebug(true)
  groups, err := api.GetGroups(false)

  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  for _, group := range groups {
    fmt.Printf("ID: %s, Name: %s\n", group.ID, group.Name)
  }
}
```

#### Getting user information

```golang
import (
  "fmt"

  "pkg.re/essentialkaos/slack.v3"
)

func main() {
  api := slack.New("YOUR_TOKEN_HERE")
  user, err := api.GetUserInfo("U023BECGF")

  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  fmt.Printf("ID: %s, Fullname: %s, Email: %s\n", user.ID, user.Profile.RealName, user.Profile.Email)
}
```

#### Minimal RTM usage

See [example](examples/websocket/websocket.go).

#### Minimal EventsAPI usage

See [example](examples/events/events.go).

### Build Status

| Branch | Status |
|------------|--------|
| `master` (_Stable_) | [![Build Status](https://travis-ci.org/essentialkaos/slack.svg?branch=master)](https://travis-ci.org/essentialkaos/slack) |
| `develop` (_Unstable_) | [![Build Status](https://travis-ci.org/essentialkaos/slack.svg?branch=develop)](https://travis-ci.org/essentialkaos/slack) |

### License

BSD 2-Clause license
