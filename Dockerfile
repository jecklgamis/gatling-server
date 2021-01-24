FROM ubuntu:20.04
MAINTAINER Jerrico Gamis <jecklgamis@gmail.com>

RUN apt update -y && apt install -y unzip openjdk-8-jre-headless
RUN rm -rf /var/lib/apt/lists/*

ENV APP_ENVIRONMENT dev

EXPOSE 8080
EXPOSE 8443

RUN mkdir -p /app/bin
RUN mkdir -p /app/configs

ENV GATLING_BUNDLE gatling-charts-highcharts-bundle-3.5.0-bundle.zip
COPY ${GATLING_BUNDLE} /app
RUN cd /app && unzip ${GATLING_BUNDLE} && rm -f ${GATLING_BUNDLE}

COPY bin/gatling-server-linux-amd64 /app/bin/gatling-server
RUN  chmod +x /app/bin/*

COPY configs /app/configs
COPY server.key /app
COPY server.crt /app

WORKDIR /app
COPY docker-entrypoint.sh /
CMD ["/docker-entrypoint.sh"]

