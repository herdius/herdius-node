FROM golang:latest  as builder
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN make build-node


# final stage
FROM scratch
COPY --from=builder /app/node /node
RUN ls
ENTRYPOINT ["/node"]
