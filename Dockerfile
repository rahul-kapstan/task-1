# Latest golang image on apline linux
FROM golang

# Work directory
WORKDIR /Task_1

# Installing dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copying all the files
COPY . .

# Starting our application
CMD ["go", "run", "main.go"]

# Exposing server port
EXPOSE 8080