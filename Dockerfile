FROM golang:1.7.3
# FROM golang:1.7.3-onbuild

# Install beego and the bee dev tool
RUN go get github.com/go-sql-driver/mysql

# Expose the application on port 4147
EXPOSE 4147

# Set the entry point of the container to the bee command that runs the
# application and watches for changes
# CMD ["bee", "run"]


# ########
# docker build -t callbacks .
# docker run -it --rm --name gcllbcks -p 4147:4147 -v /home/ekt/go/src/gcllbcks:/go/src/gcllbcks -w /go/src/gcllbcks callbacks
