FROM oraclelinux:8

ARG arg_dbuser="admin"
ARG arg_dbpassword
ARG arg_dbhost="db.mplinksters.club:3306"
ARG	arg_dbname="mplinksters"
ARG arg_listenport="8000"

ENV dbuser=${arg_dbuser}
ENV dbpassword=${arg_dbpassword}
ENV dbhost=${arg_dbhost}
ENV dbname=${arg_dbname} 
ENV listenport=${arg_listenport}

EXPOSE ${listenport}

COPY mpapiserver /mpapiserver

ENTRYPOINT [ "/mpapiserver" ]



