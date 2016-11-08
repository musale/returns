package common

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

var DbCon *sql.DB

