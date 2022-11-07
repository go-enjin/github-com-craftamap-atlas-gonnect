# atlas-gonnect : An Atlassian Connect Framework written in Golang

## Notice

This project is a fork of [github.com/craftamap/atlas-gonnect](https://github.com/craftamap/atlas-gonnect).

There are a few motivations for this fork:

- use [chi](https://github.com/go-chi/chi) instead of [mux](https://github.com/gorilla/mux)
- use the latest version of [gorm](https://gorm.io)
- handle logging with `github.com/go-enjin/be/pkg/log`
- use structs to create new addons instead of .json files
- any changes necessary to support the atlassian Go-Enjin feature

Note that the rest of this README has been trimmed to basics as this package
is not intended to be used directly. Please see the third-party Go-Enjin
[atlassian](https://github.com/go-enjin/third_party/blob/trunk/features/atlassian)
feature instead.

This project is not associated with Atlassian.

## Upstream Author

Fabian Siegel

## License

Apache 2.0.
