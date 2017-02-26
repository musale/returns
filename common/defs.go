package common

import (
	"database/sql"
	"log"
)

//failed constant
const failed string = "failed"

// POST method constant
const POST string = "POST"

var DbCon *sql.DB

var Logger *log.Logger
