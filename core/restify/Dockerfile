FROM scala

ENV DEBIAN_FRONTEND noninteractive

#
# Copy app jars
COPY build/distributions/restify.tar ./
RUN tar xf restify.tar
CMD restify/bin/restify >> /logs/${COMPONENT_NAME}_logs.log 2>&1

EXPOSE 8080

