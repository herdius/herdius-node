FROM golang:latest 
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN make build-node
CMD ["/app/node -port=3001 " ]