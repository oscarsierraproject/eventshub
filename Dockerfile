FROM debian:bookworm-slim

LABEL maintainer "Sebastian Oleksiak <oscarsierraproject@protonmail.com>"

ENV APP_PATH="/app"

RUN mkdir -p ${APP_PATH}/
COPY ./bin/main ${APP_PATH}/eventshub
RUN chmod +x ${APP_PATH}/eventshub

ENTRYPOINT ["/app/eventshub"]