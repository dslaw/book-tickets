# syntax-docker/dockerfile:1

FROM python:3.13-alpine

WORKDIR /usr/src/pgsync

RUN apk add --no-cache bash

RUN pip install pgsync==3.3.0 sqlalchemy==2.0.36

COPY schema.json run.sh ./

CMD ["/bin/bash", "run.sh"]
