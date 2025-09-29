# go-inceptiondb

Go client for the [InceptionDB](https://inceptiondb.hola.cloud) REST API.

## Installation

```
go get github.com/holaql/go-inceptiondb
```

## Usage

Create a client by providing the target database identifier and API
credentials. The base URL defaults to the SaaS endpoint, but you can override
it when self-hosting:

```go
client, err := inceptiondb.NewClient(inceptiondb.Config{
        DatabaseID: "89744b21-5ab6-42a1-9209-3805c9c66834",
        APIKey:     "<api-key>",
        APISecret:  "<api-secret>",
})
if err != nil {
        log.Fatal(err)
}
```

Credentials are automatically sent on every request, so you only need to focus
on the operations you want to perform. The client exposes helpers to manage
collections, indexes and documents using idiomatic Go types.

### Example project

The `examples/todo` folder contains a small command-line application that uses a
real database to manage a to-do list. Configure the required environment
variables and run it with:

```
export INCEPTIONDB_DATABASE_ID="<database-id>"
export INCEPTIONDB_API_KEY="<api-key>"
export INCEPTIONDB_API_SECRET="<api-secret>"
go run ./examples/todo
```

The program creates a collection, inserts a few tasks, reads back the results in
pages and finally drops the collection so you can re-run it safely.

### Testing

Run the library tests with:

```
go test ./...
```
