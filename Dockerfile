# syntax=docker/dockerfile:1

##
## STEP 1 - BUILD
##

# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.19-alpine AS build

# create a working directory inside the image
WORKDIR /app

# copy Go modules and dependencies to image
COPY go.mod go.sum ./

# download Go modules and dependencies
RUN go mod download

# copy directory files i.e all files ending with .go (+ version-tag)
COPY *.go version.txt ./

# compile application (static linked)
RUN CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o /melp

##
## STEP 2 - DEPLOY
##
FROM scratch

ENV PATH=/

WORKDIR /

COPY --from=build /melp /

EXPOSE 10000

CMD ["/melp"]