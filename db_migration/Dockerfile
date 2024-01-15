FROM ubuntu:22.04

WORKDIR /home/app
RUN apt-get update
RUN apt-get install -y curl
RUN apt-get install -y gnupg
RUN curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
RUN  echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ xenial main" > /etc/apt/sources.list.d/migrate.list
RUN apt-get update
RUN apt-get install -y migrate

COPY ./migrate.sh .
COPY /migrations migrations/
RUN chmod +x ./migrate.sh
