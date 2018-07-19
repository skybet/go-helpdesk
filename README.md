# go-helpdesk

## A library for building helpdesk bots in Go

[![CircleCI](https://circleci.com/gh/skybet/go-helpdesk/tree/master.svg?style=svg)](https://circleci.com/gh/skybet/go-helpdesk/tree/master) [![CodeFactor](https://www.codefactor.io/repository/github/skybet/go-helpdesk/badge)](https://www.codefactor.io/repository/github/skybet/go-helpdesk) [![Coverage Status](https://coveralls.io/repos/github/skybet/go-helpdesk/badge.svg?branch=master)](https://coveralls.io/github/skybet/go-helpdesk?branch=master)

`go-helpdesk` is a library and standalone binary designed to be run as a server to respond to Slack slash commands, although it should be possible to use any messaging platform with little effort. Once a user has engaged with the app, you can then configure it to raise JIRA tickets, call someone out on PagerDuty or whatever. The library consists of a server which responds to webhooks from messaging platforms and some common handlers and a wrapper library for Slack and JIRA. It is a simple aid to take some of the sting out of first-line support and incident management using ChatOps.

It is possible to use the basic bundled functionality by simply deploying the bundled implementation or you can import it as a library and write your own.

## Development

### Dependencies

We have chosen to commit the vendor tree as it gives us repeatable builds and control over our dependencies. Please run `dep ensure` to update the vendor tree. This is not done by CI.

### Automated Testing

We use [Mockery](https://github.com/vektra/mockery) for test mocks. Use `go generate` to regenerate test doubles after you have go getted the mockery package. If you need to add new interfaces to be mocked, add new generate comments to the top of `main.go`.

Any commit without appropriate test coverage will be rejected.

## CLI Usage

### Flags

```
  -a, --app-token string        Slack API token for your slash command (required)
  -b, --bot-token string        Slack API token for bot integration (required)
  -l, --listen-address string   Address to listen for Slack callbacks on (default ":4390")
```

### Environment Variables

You can set flags from environment variables instead. You simply take the log form of the flag and prefix it with `HELP_`, replacing any hyphens with underscores. 

> Example: `HELP_APP_TOKEN` will set the `app-token` flag

_Nb._ Flags take precedence over environment variables.

### Slack Tokens

`go-helpdesk` requires two different tokens to connect to Slack. An app token is provided when creating a new slash command and a bot token is required to send messages etc. (_TODO: expand this_)

## Library Usage

Check the example `main.go` (_TODO: write a proper guide once API is stable_)
