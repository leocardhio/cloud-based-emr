FROM mongodb/mongodb-enterprise-server:latest

USER root

RUN wget https://go.dev/dl/go1.20.6.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.6.linux-amd64.tar.gz && rm go1.20.6.linux-amd64.tar.gz
RUN apt-get update && apt-get install -y curl pkg-config
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin


WORKDIR /app/
COPY . /app/

RUN sh -c 'curl -s --location https://www.mongodb.org/static/pgp/libmongocrypt.asc | gpg --dearmor >/etc/apt/trusted.gpg.d/libmongocrypt.gpg'
RUN echo "deb https://libmongocrypt.s3.amazonaws.com/apt/ubuntu jammy/libmongocrypt/1.8 universe" | tee /etc/apt/sources.list.d/libmongocrypt.list
RUN apt-get update && apt-get install -y libmongocrypt-dev

RUN go install
RUN go env -w CGO_ENABLED=1 && go build --tags cse

EXPOSE 8081

CMD ./service-lab