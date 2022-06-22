# oaimerge
oaimerge loads an openapi yaml specification (eventually from multiple files), 
applies a variable interpolation and merges everything into a single openapi json.
Additional checks or linting rules are not applied.

## why would one ever need that?

For example, if you want to use external descriptions (e.g. because you have used [xtractdoc](https://github.com/worldiety/xtractdoc)),
there is currently no support from the OpenAPI specification. See also the following related tickets:

* https://github.com/OAI/OpenAPI-Specification/issues/2697
* https://github.com/OAI/OpenAPI-Specification/issues/556
* https://github.com/OAI/OpenAPI-Specification/issues/1514

See also [RFC3986](https://datatracker.ietf.org/doc/html/rfc3986). 
In contrast to the [official OAI spec](https://swagger.io/docs/specification/using-ref/), _$ref_ is allowed anywhere,
only limited by the yaml grammar.

## usage

```bash
go install github.com/worldiety/oaimerge/cmd/oaimerge@latest
oaimerge -oai=openapi.yaml > mergedoai.json
```

## limitations

* supports only local relative yaml files. There is no generic URI or URL support
* string interpolation works only for inline strings of the form `$ref{filename#/wdy.de-my/jsonptr}`