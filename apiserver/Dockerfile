FROM oraclelinux:8-slim

ARG arg_dbuser="admin"
ARG arg_dbpassword="WD37wn^R##jG"
ARG arg_dbhost="db.mplinksters.club:3306"
ARG	arg_dbname="mplinksters"

ENV dbuser=${arg_dbuser}
ENV dbpassword=${arg_dbpassword}
ENV dbhost=${arg_dbhost}
ENV dbname=${arg_dbname} 

COPY mpapiserver /mpapiserver

ENTRYPOINT [ "/mpapiserver" ]



