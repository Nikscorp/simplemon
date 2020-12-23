FROM nikscorp/go-builder:0.1.2 as build-backend

ADD app /go/src/simplemon/app/
ADD vendor /go/src/simplemon/vendor/
ADD conf /go/src/simplemon/conf
ADD .golangci.yml /go/src/simplemon

WORKDIR /go/src/simplemon
RUN go build -o simplemon simplemon/app
RUN golangci-lint run ./...

FROM alpine:3.12.1

COPY --from=build-backend /go/src/simplemon/simplemon /srv/simplemon
COPY --from=build-backend /go/src/simplemon/conf/simplemon-conf.yml /etc/simplemon/simplemon.yml

CMD /srv/simplemon -c /etc/simplemon/simplemon.yaml
