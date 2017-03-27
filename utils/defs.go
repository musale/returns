package utils

import (
	"database/sql"

	"github.com/garyburd/redigo/redis"
)

//failed constant
const failed string = "failed"

// POST method constant
const POST string = "POST"

// DBCon connection to mysql
var DBCon *sql.DB

// RedisCon common redis object
var RedisCon *redis.Pool
