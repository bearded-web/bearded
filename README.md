# Bearded (WIP)

[![travis](https://travis-ci.org/bearded-web/bearded.svg)](https://travis-ci.org/bearded-web/bearded)  [![Coverage Status](https://coveralls.io/repos/bearded-web/bearded/badge.svg?branch=develop)](https://coveralls.io/r/bearded-web/bearded?branch=develop)

[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/bearded-web/bearded?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)


Bearded is an open source Security Automation platform. The platform allows Development, QA, and Security team members to perform automated web security scans with a set of tools, and re-execute those scans as needed. All tools can be executed in the cloud in docker containers.  Barbudo has a default web interface which integrates all core options and makes it possible to manage large pentests easily.

### Screenshots and Video

Demo video:
 - English: https://www.youtube.com/watch?v=i--w-fD_uSo
 - Russian: https://www.youtube.com/watch?v=GY58-dvF9Og

Interface:
 - http://barbudo.net/img/screens/feed.png
 - http://barbudo.net/img/screens/scan.png
 - http://barbudo.net/img/screens/report.png

__The project in pre-pre-alpha stage. Please, do not use this right now.__




### dev
Update
`go get -u -d github.com/bearded-web/bearded`


Go to path
`cd $GOPATH/src/github.com/bearded-web/bearded`

Update dependencies
`godep restore`

Build
`go get github.com/bearded-web/bearded`

Run dev server and look to build path
`bearded dispatcher --frontend ../frontend/build/ -v`

In `../frontend/` exec `npm run dev` to start frontend static server

Load data:

`bearded plugins load --update ./extra/data/plugins.json`

`bearded plans load --update ./extra/data/plans.json`

Swagger `http://127.0.0.1:3003/apidocs/`

# local dev
Make changes, then
`go get github.com/bearded-web/bearded`
