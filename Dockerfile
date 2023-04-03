FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /voinc-backend

EXPOSE 8080

CMD [ "/voinc-backend" ]