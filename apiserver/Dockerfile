FROM oraclelinux:8

ARG arg_listenport="8000"

ENV listenport=${arg_listenport}

EXPOSE ${listenport}

COPY mpapiserver /mpapiserver

ENTRYPOINT [ "/mpapiserver" ]



