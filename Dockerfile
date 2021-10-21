FROM golang:1.15 as build
WORKDIR /go/src/github.com/universityofadelaide/openshift-newrelic-synthetics
ADD . /go/src/github.com/universityofadelaide/openshift-newrelic-synthetics
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/openshift-newrelic-synthetics github.com/universityofadelaide/openshift-newrelic-synthetics/cmd/openshift-newrelic-synthetics

FROM alpine:3.10
RUN apk --no-cache add ca-certificates
COPY --from=build /go/src/github.com/universityofadelaide/openshift-newrelic-synthetics/bin/openshift-newrelic-synthetics /usr/local/bin/openshift-newrelic-synthetics
ENTRYPOINT ["openshift-newrelic-synthetics"]
