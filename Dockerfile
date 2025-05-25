FROM registry.fedoraproject.org/fedora-minimal:40-x86_64

EXPOSE 8080/tcp

LABEL org.opencontainers.image.source https://github.com/jamesread/orgviz
LABEL org.opencontainers.image.title orgviz

COPY frontend-dist /usr/share/orgviz/frontend/
COPY var/config-skel/ /config/
COPY orgviz /app/orgviz

RUN mkdir -p /config/exec/

VOLUME /config
VOLUME /usr/libexec/orgviz/

ENTRYPOINT [ "/app/orgviz" ]
