package utils

import (
	"database/sql"
)

//failed constant
const FAILED string = "failed"

// POST method constant
const POST string = "POST"

// DBCon connection to mysql
var DBCon *sql.DB
