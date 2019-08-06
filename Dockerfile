FROM golang:latest  as builder
RUN mkdir /app 
ADD . /app/ 

WORKDIR /app 
RUN ls
RUN pwd
RUN make build-node
RUN ls
RUN ls /app
ENTRYPOINT ["/app/node"]





# final stage
# FROM alpine as built
# COPY --from=builder /app/ /app
# RUN ls /app
# RUN pwd
# RUN ls /app
