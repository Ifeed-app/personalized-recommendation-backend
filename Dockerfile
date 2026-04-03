FROM ubuntu:latest
LABEL authors="dario"

ENTRYPOINT ["top", "-b"]