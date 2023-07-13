FROM redis:alpine

ARG CONSUL_VERSION

RUN mkdir -p /etc/consul.d/check_scripts
COPY ./redis-instance-check.sh /etc/consul.d/check_scripts
RUN chmod +x /etc/consul.d/check_scripts/redis-instance-check.sh

RUN mkdir -p /etc/consul.d/client
COPY ./redis-consul-config.json /etc/consul.d/client

ADD https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_amd64.zip /tmp/consul.zip
RUN cd /bin && unzip /tmp/consul.zip && chmod +x /bin/consul && rm /tmp/consul.zip

COPY entrypoint.sh /bin
RUN chmod +x /bin/entrypoint.sh

CMD ["/bin/entrypoint.sh"]