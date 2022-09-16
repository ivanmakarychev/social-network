# syntax=docker/dockerfile:1

FROM centos:centos7

RUN yum -y install git ncurses-devel cmake gcc-c++ boost boost-devel wget unzip nano bzip2 mysql-devel mysql-lib

RUN git clone https://github.com/tarantool/mysql-tarantool-replication.git
WORKDIR mysql-tarantool-replication
RUN git submodule update --init --recursive
RUN cmake .
RUN make

COPY replicatord.yml ./replicatord.yml

CMD ["replicatord", "-c", "./replicatord.yml"]