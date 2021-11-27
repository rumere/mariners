FROM oraclelinux:8

RUN dnf -y install mysql mysql-server
RUN dnf -y update && \
    dnf clean all && \
    rm -rf /var/cache/yum
