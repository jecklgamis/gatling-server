FROM ubuntu:20.04
MAINTAINER Jerrico Gamis <jecklgamis@gmail.com>

RUN apt update -y && apt install -y unzip openjdk-8-jdk-headless curl dumb-init
RUN rm -rf /var/lib/apt/lists/*

ENV APP_ENVIRONMENT dev

EXPOSE 8080
EXPOSE 8443

RUN mkdir -p /app/bin
RUN mkdir -p /app/configs

ARG GATLING_VERSION=3.9.5
ENV GATLING_BUNDLE="gatling-charts-highcharts-bundle-${GATLING_VERSION}"
ENV GATLING_BUNDLE_ZIP="${GATLING_BUNDLE}-bundle.zip"
ENV GATLING_DOWNLOAD_URL=https://repo1.maven.org/maven2/io/gatling/highcharts/gatling-charts-highcharts-bundle/${GATLING_VERSION}/${GATLING_BUNDLE_ZIP}

WORKDIR /app
RUN curl -fLO ${GATLING_DOWNLOAD_URL} && unzip ${GATLING_BUNDLE_ZIP} && rm -f ${GATLING_BUNDLE_ZIP}

COPY scripts/gatling-runner.sh /app/${GATLING_BUNDLE}/bin/
COPY scripts/gatling-jar-runner.sh /app/${GATLING_BUNDLE}/bin/

COPY bin/gatling-server-linux-amd64 /app/bin/gatling-server
RUN  chmod +x /app/bin/*

COPY configs /app/configs
COPY server.key /app
COPY server.crt /app

RUN groupadd app && useradd -g app app -m -d /home/app
RUN chown -R  app:app /app
USER app

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

COPY docker-entrypoint.sh /
CMD ["/docker-entrypoint.sh"]

