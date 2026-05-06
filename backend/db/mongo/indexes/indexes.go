package indexes

import _ "embed"

// BooksIndexesJSON contains the raw JSON data for MongoDB indexes.
//
//go:embed books_indexes.json
var BooksIndexesJSON []byte
