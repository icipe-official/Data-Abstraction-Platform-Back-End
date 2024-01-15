FROM debian:bookworm-slim

WORKDIR /home/app

RUN apt-get update
RUN apt-get install ca-certificates -y
RUN update-ca-certificates

COPY /bin .
RUN mkdir tmp

RUN chmod +x create_system_user
RUN chmod +x init_database
RUN chmod +x api

EXPOSE 3000

CMD [ "./api" ]