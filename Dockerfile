FROM golang:1.13

# Set the home directory to ~
ENV HOME ~
# cd into the home directory
WORKDIR ~

COPY . .

# Allow port 8000 to be accessed
# from outside the container
EXPOSE 8000

# Grab MongoDB dependency for golang
# RUN go get go.mongodb.org/mongo-driver/mongo

# Wait for db
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.2.1/wait /wait
RUN chmod +x /wait

# Run the app
CMD /wait && go run server.go