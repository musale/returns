package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/etowett/returns/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Code short code struct
type Code struct {
	CodeID   int64
	CodeType string
	UserID   int64
	Code     string
}

// SharedCode dat
type SharedCode struct {
	UserID  int64
	Keyword string
}

// AutoResponse data
type AutoResponse struct {
	SenderID string
	Keyword  string
	Message  string
	UserID   string
}

func main() {
	err := godotenv.Load("/apps/returns/.env")
	if err != nil {
		log.Fatal("Error loading .env file ", err)
	}

	dbObj, err := sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASS")+"@tcp("+os.Getenv("DB_HOST")+":3306)/"+os.Getenv("DB_NAME")+"?charset=utf8")
	if err != nil {
		log.Fatal("db error: ", err)
	}
	defer dbObj.Close()

	// Test the connection to the database
	err = dbObj.Ping()
	if err != nil {
		log.Fatal("Error DB ping ", err)
	}

	redisCon := utils.RedisPool().Get()
	defer redisCon.Close()

	var codes []Code

	stmt, err := dbObj.Prepare("select id, code_type, user_id, code from callbacks_code")

	rows, err := stmt.Query()

	if err != nil {
		log.Fatal("query select out", err)
	}

	for rows.Next() {
		var code Code
		var userID sql.NullInt64
		err := rows.Scan(&code.CodeID, &code.CodeType, &userID, &code.Code)
		if err != nil {
			log.Fatal("error scan out", err)
		}
		if userID.Valid {
			code.UserID = userID.Int64
		}
		codes = append(codes, code)
	}

	for _, code := range codes {
		codeString := "code:" + code.Code
		if code.CodeType == "DEDICATED" {
			redisCon.Do("HMSET", codeString, "user_id", code.UserID, "code_type", "DEDICATED")
		} else {
			redisCon.Do("HMSET", codeString, "code_type", "SHARED")

			var sharedDetails []SharedCode

			stmt, err := dbObj.Prepare("select user_id, keyword from callbacks_shared where code_id=?")

			rows, err := stmt.Query(code.CodeID)

			if err != nil {
				log.Fatal("query select out", err)
			}

			for rows.Next() {
				var shared SharedCode
				err := rows.Scan(&shared.UserID, &shared.Keyword)
				if err != nil {
					log.Fatal("error scan out", err)
				}
				sharedDetails = append(sharedDetails, shared)
			}
			for _, sha := range sharedDetails {
				redisCon.Do("HMSET", codeString, sha.Keyword, sha.UserID)
			}
		}
	}

	stmt, err = dbObj.Prepare(
		"select senderid, `key`, message, user_id from callbacks_autoresponse",
	)

	rows, err = stmt.Query()

	if err != nil {
		log.Fatal("query select out", err)
	}

	var auto []AutoResponse
	for rows.Next() {
		var aut AutoResponse
		err = rows.Scan(
			&aut.SenderID, &aut.Keyword, &aut.Message, &aut.UserID,
		)
		if err != nil {
			log.Fatal("error scan out", err)
		}

		auto = append(auto, aut)
	}

	for _, aut := range auto {
		keyString := "auto:" + aut.Keyword + ":" + aut.UserID
		keyVal := map[string]string{
			"message": aut.Message, "sender_id": aut.SenderID,
		}
		jsonString, err := json.Marshal(keyVal)
		if err != nil {
			log.Fatal("error json marshal: ", err)
		}
		if _, err = redisCon.Do("SET", keyString, jsonString); err != nil {
			log.Fatal("Redis set error: ", err)
		}
	}

	log.Println("Redis populated successfully")
	return
}
