ARG CONNECT_VERSION=7.9.1

FROM alpine:3.20 as debezium
WORKDIR /tmp
RUN mkdir /jars
RUN wget https://repo1.maven.org/maven2/io/debezium/debezium-connector-mongodb/3.1.1.Final/debezium-connector-mongodb-3.1.1.Final-plugin.tar.gz
RUN tar -xvf debezium-connector-mongodb-*.tar.gz
RUN mv -f debezium-connector-mongodb/*.jar /jars

FROM confluentinc/cp-kafka-connect:${CONNECT_VERSION}
COPY --from=debezium /jars /usr/share/java/kafka
RUN echo installing plugins...\
&& confluent-hub install confluentinc/kafka-connect-avro-converter:latest --no-prompt \
&& confluent-hub install confluentinc/kafka-connect-s3:latest --no-prompt

