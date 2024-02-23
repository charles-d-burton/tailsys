FROM public.ecr.aws/docker/library/golang:alpine3.19 as base

RUN apk update && apk add --no-cache git
WORKDIR /app

COPY . .

RUN go build -o tailsys

FROM scratch as coordination

WORKDIR /app
COPY --from=base /app/tailsys /app/tailsys

ENTRYPOINT [ "/app/tailsys" ]


FROM scratch as client

WORKDIR /app
COPY --from=base /app/tailsys /app/tailsys

ENTRYPOINT [ "/app/tailsys" ]

