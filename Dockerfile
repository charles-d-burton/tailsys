FROM public.ecr.aws/docker/library/golang:alpine3.19 AS base

RUN apk update && apk add --no-cache git
WORKDIR /app

COPY . .

RUN go build -o tailsys

####################################
# Coordination server
####################################
FROM public.ecr.aws/h1a5s9h8/alpine:latest AS coordination

WORKDIR /app
COPY --from=base /app/tailsys /app/tailsys

ENTRYPOINT [ "/app/tailsys"]


####################################
# Client system to test build
####################################
FROM public.ecr.aws/h1a5s9h8/alpine:latest AS client

WORKDIR /app
COPY --from=base /app/tailsys /app/tailsys

ENTRYPOINT [ "/app/tailsys"]


# ####################################
# # Non-interactive test container
# ####################################
# FROM scratch as ni

# WORKDIR /app
# COPY --from=base /app/tailsys /app/tailsys
#COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs

# ENTRYPOINT [ "/app/tailsys"]
