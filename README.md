# Bearded WIP


Work in progress, do not use this


### dev
Update
`go get -u github.com/bearded-web/bearded`

Go to path
`cd $GOPATH/src/github.com/bearded-web/bearded`

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