### Callbacks Handler
> This simple service was created to handle the callbacks from APIs. These callbacks are bombarded to the service, stored in Redis and then updated in a blocking manner in the MySQL db.

### Dependencies
* `"github.com/etowett/returns/common"` -  contains all the common methods used in `"github.com/etowett/returns/mylib`
* `"github.com/garyburd/redigo/redis"` -  Golang client for redis
* `"github.com/go-sql-driver/mysql"` - Golang driver for MySQL
* `github.com/joho/godotenv` - Golang loader for environment variables from `.env`
> Most of the other dependencies are found in the import statements

### Getting Started
###### Description
* `main.go` is the main file that is run for execution i.e. `go run main.go`
* The `common` folder contains the common methods re-used throughout the project.
* The `config` folder houses configuration for use if you want to run this application as a service with `systemd`
* `mylib` contains the `structs`, `interfaces` and `functions` that handle the receiving of callbacks, parsing into required data structures and processing into the final step.

###### Setting up
* Ensure you've setup `Golang` and your `$GO_PATH` (if not, please Google :smirk: )
* Do `go get github.com/etowett/returns`. This will clone a copy of this repo into your `/path/to/golang/projects/src/github.com/etowett`
###### .env and environment variables
> Rename the `env.sample` file into `.env` and fill out the required variables

###### Redis
> Redis should be running locally.
`github.com/etowett/returns/common/utils.go` has the following function to initialize a Redis pool.
```
    // RedisPool returns a redis pool
    func RedisPool() *redis.Pool {
        return &redis.Pool{
            MaxIdle:   80,
            MaxActive: 12000, // max number of connections
            Dial: func() (redis.Conn, error) {
                c, err := redis.Dial("tcp", ":6379")
                if err != nil {
                    panic(err.Error())
                }
                return c, err
            },
        }
    }
```
The connection is pretty normal

###### MySQL
> There are 2 scenarios for connection in MySQL. We use a remote DB so our connection looks like this in `github.com/etowett/returns/main.go`:
```
    common.DbCon, err = sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASS")+"@tcp("+os.Getenv("DB_HOST")+":3306)/"+os.Getenv("DB_NAME")+"?charset=utf8")
	if err != nil {
		panic(err.Error())
	}
	defer common.DbCon.Close()
```

> If your DB is in your local server, it would be something like this:
```
common.DbCon, err = sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASS")+os.Getenv("DB_NAME")+"?charset=utf8")
if err != nil {
    panic(err.Error())
}
defer common.DbCon.Close()
```
Just remove the `tcp` part.

###### Logging
* Create a folder for the log file as specified in your `LOG_DIR` variable in `.env`
* Open a terminal and `tail -f /path/to/logs/returns.log`
> Ensure that this folder has the proper permissions to allow writing into it's files

###### Building and installing
* Run `go build` to build the app
* Run `go install`. Check the `/path/to/golang/projects/bin` folder and you should find a `returns` executable.
* Run `./returns` on one terminal.
* Open another terminal and run the python requests simulator file `client.py` i.e. `./client.py`
> You should see outputs of `Dlrs received` and logs on the tailed logs terminal.

### TODOs
* Dockerize the app so that it runs in it's own container
* Handle callbacks for `inboxes`
* Handle callbacks for `opt-out`

### Contributing
> Contributions are welcome. PRs will be accepted if they've followed atleast the minimum standards.

### Issues
> Reach out to:
* [Eutychus Towett](https://github.com/etowett)
* [Martin Musale](https://github.com/musale)
