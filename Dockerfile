FROM ubuntu:24.04
LABEL org.opencontainers.image.authors="jecklgamis@gmail.com"


RUN apt update -y && apt install -y openjdk-21-jdk-headless curl dumb-init
RUN rm -rf /var/lib/apt/lists/*

ENV APP_ENVIRONMENT=dev

EXPOSE 8080
EXPOSE 8443

RUN mkdir -p /app/bin
RUN mkdir -p /app/configs
RUN mkdir -p /app/scripts

WORKDIR /app

COPY scripts/gatling-jar-runner.sh /app/scripts/

COPY bin/gatling-server-linux-amd64 /app/bin/gatling-server
RUN  chmod +x /app/bin/* /app/scripts/*

COPY configs /app/configs
COPY server.key /app
COPY server.crt /app

RUN groupadd app && useradd -g app app -m -d /home/app
RUN chown -R  app:app /app
USER app

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

COPY docker-entrypoint.sh /
CMD ["/docker-entrypoint.sh"]

