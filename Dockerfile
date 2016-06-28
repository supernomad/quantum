FROM golang:1.6

ENV WORKING "/usr/local/go/src/github.com/Supernomad/quantum"

RUN mkdir -p $WORKING

COPY $PWD $WORKING

RUN cd $WORKING \
    && go get -d -v \
    && go build -v \
    && chmod +x quantum \
    && mv quantum /bin/quantum

ENTRYPOINT ["quantum"]
