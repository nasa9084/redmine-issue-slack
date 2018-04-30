# redmine-issue-slack
=====================

a slack bot searches your redmine ticket by ticket ID

![screenshot](redmine-issue-slack-screenshot.png)

## INSTALL

``` shell
$ go get -u github.com/nasa9084/redmine-issue-slack
```

## USAGE

``` shell
$ redmine-issue-slack [-t SLACK_TOKEN] [-r REDMINE_ENDPOINT] [-k REDMINE_APIKEY]
```

### ARGs

| short | long               | env              | description         |
|:-----:|:------------------:|:----------------:|:-------------------:|
| -t    | --slack-token      | SLACK_TOKEN      | API Token for slack |
| -r    | --redmine-endpoint | REDMINE_ENDPOINT | Endpoint of redmine |
| -k    | --redmine-apikey   | REDMINE_APIKEY   | API key for redmine |
