# Build FrontEnd
FROM node:alpine AS frontend
WORKDIR /frontend/
COPY ./frontend/src ./src
COPY ./frontend/favicon.svg ./favicon.svg
COPY ./frontend/README.md ./README.md
COPY ./frontend/index.html ./index.html
COPY ./frontend/package-lock.json ./package-lock.json
COPY ./frontend/.eslintrc ./.eslintrc
COPY ./frontend/jsconfig.json ./jsconfig.json
COPY ./frontend/package.json ./package.json
COPY ./frontend/vite.config.js ./vite.config.js
RUN npm install -D vite
RUN npm install
RUN npm run build

# Build
FROM golang:alpine AS build
WORKDIR /build/
COPY go.mod .
RUN go mod tidy
COPY . .
COPY --from=frontend /frontend/dist/assets ./assets
COPY --from=frontend ./frontend/dist/index.html ./
RUN go version
RUN go build

# Production
FROM alpine AS production
WORKDIR /chrdsuimanager/

RUN apk add --upgrade --no-cache ca-certificates && update-ca-certificates

COPY --from=build /build/chrdsuimanager ./
COPY ./*.crt ./
COPY ./*.key ./
COPY ./*.pem ./
COPY ./chrdsuimanager.json ./

EXPOSE 7007

CMD ["./chrdsuimanager", "-log=stdout"]
