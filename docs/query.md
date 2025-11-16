# Query

> [!WARNING]
>
> Very much work-in-progress

The heart of skabelon is the http-to-sql query model. Inspired by the interfaces
in other popular tools:

* [postgREST](https://docs.postgrest.org/en/v12/references/api/tables_views.html)
* [Pocketbase](https://pocketbase.io/docs/api-records/)
* [pRESTd](https://docs.prestd.com/api-reference/advanced-queries)

Also, the proposed
[HTTP QUERY method](https://httpwg.org/http-extensions/draft-ietf-httpbis-safe-method-w-body.html)
is interesting.

A `http.Request` with query parameters or form data contain
[url.Values](https://pkg.go.dev/net/url#Values).

```go
type Values map[string][]string
```


## pRESTd

* `<field name>`
* `_select` - etc

Good idea: underscore prefix!

Groupby?

## Pocketbase

Query parameters on `List`.

* `filter` - Filter expression to filter/search the returned records list
* `fields` - Comma separated string of the fields to return in the JSON response
  (by default returns all fields).
* `expand` - Auto expand record relations.


## ID? Name vs int vs uuid v7?

`/my/some-name` vs `/my/12314` vs `/my/123f-42415-125-123`

## Dataprovider operations

* get - by id
* list - w/ filter
* create
* update
* delete

And `many`-endpoints for:

* get - `in.(1,2,5,6)`
* create - list instead of dict
* update - `in.(1,2,5,6)`
* delete - `in.(1,2,5,6)`

## OR vs AND

## Paths

* `GET /resource?id=eq.1`
