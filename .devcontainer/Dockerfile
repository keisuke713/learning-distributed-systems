FROM ubuntu:24.04

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        openjdk-21-jdk \
        git \
        openssh-client \
        curl \
    && rm -rf /var/lib/apt/lists/*

# is needed?
# RUN apt-get update && \
#     apt-get install -y wget && \
#     rm -rf /var/lib/apt/lists/*

RUN curl -fL https://github.com/coursier/coursier/releases/latest/download/cs-x86_64-pc-linux.gz | gzip -d > cs && \
    chmod +x cs && ./cs setup -y

SHELL ["/bin/bash", "-c"]
RUN source ~/.profile

RUN  apt-get update && \
     apt-get -y install apt-transport-https gnupg -yqq && \
     echo "deb https://repo.scala-sbt.org/scalasbt/debian all main" | tee /etc/apt/sources.list.d/sbt.list && \
     echo "deb https://repo.scala-sbt.org/scalasbt/debian /" | tee /etc/apt/sources.list.d/sbt_old.list && \
     curl -sL "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x2EE0EA64E40A89B84B2DF73499E82A75642AC823" | gpg --no-default-keyring --keyring gnupg-ring:/etc/apt/trusted.gpg.d/scalasbt-release.gpg --import && \
     chmod 644 /etc/apt/trusted.gpg.d/scalasbt-release.gpg && \
     apt-get update && \
     apt-get install -y sbt

