FROM golang:1.16.4 as builder

RUN go env -w GOPROXY=https://goproxy.cn
COPY . /go/src/app

# build app
# git describe --tags `git rev-list --tags --max-count=1` > VERSION && \
RUN cd /go/src/app && \
	date "+%Y-%d-%mT%TZ%z" > DATE && \
	echo "DATE:" $(cat DATE) && \
	git  branch | grep '*' | sed -e 's/\*//g' -e 's/HEAD detached at//g' -e 's/\s*//g' -e 's/[\(\)]//g'  > VERSION && \
	echo "VERSION:" $(cat VERSION) && \
	CGO_ENABLED=0 go build --ldflags "$LDFLAGS -s -w" -a -installsuffix cgo -v -o /app


FROM alpine:3.13.5

ARG UID=10000
ARG GID=10000
ARG ADDITIONAL_PACKAGE

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.13/main" > /etc/apk/repositories && \
	addgroup -g $GID -S app && adduser -u $UID -S -g app app && \
	apk --no-cache add tzdata ${ADDITIONAL_PACKAGE}

COPY --from=builder /app /app/entry
WORKDIR /app
USER app
ENV TZ=Asia/Shanghai
EXPOSE 8080

CMD [ "/app/entry" ]
