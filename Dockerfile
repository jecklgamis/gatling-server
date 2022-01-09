FROM ubuntu:20.04
MAINTAINER Jerrico Gamis <jecklgamis@gmail.com>

RUN apt update -y && apt install -y unzip openjdk-8-jdk-headless curl dumb-init
RUN rm -rf /var/lib/apt/lists/*

ENV APP_ENVIRONMENT dev

EXPOSE 8080
EXPOSE 8443

RUN mkdir -p /app/bin
RUN mkdir -p /app/configs

ENV GATLING_BUNDLE gatling-charts-highcharts-bundle-3.7.3-bundle.zip
ENV GATLING_DIST_DIR gatling-charts-highcharts-bundle-3.7.3
COPY ${GATLING_BUNDLE} /app
RUN cd /app && unzip ${GATLING_BUNDLE} && rm -f ${GATLING_BUNDLE}
COPY scripts/gatling-v3.7.3.sh /app/${GATLING_DIST_DIR}/bin/gatling.sh

COPY bin/gatling-server-linux-amd64 /app/bin/gatling-server
RUN  chmod +x /app/bin/*

COPY configs /app/configs
COPY server.key /app
COPY server.crt /app

RUN groupadd app && useradd -g app app -m -d /home/app
RUN chown -R  app:app /app
USER app

WORKDIR /app
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

COPY docker-entrypoint.sh /
CMD ["/docker-entrypoint.sh"]

