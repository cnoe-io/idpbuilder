FROM nginx:stable

RUN set -x && \
  apt-get -y update                                     &&  \
  apt-get install -y git fcgiwrap spawn-fcgi wget       &&  \

  adduser git --home /var/lib/git                       &&  \
  adduser nginx git                                     &&  \

  git config --system http.receivepack true             &&  \
  git config --system http.uploadpack true              &&  \
  git config --system user.email "gitserver@git.com"    &&  \
  git config --system user.name "Git Server"            &&  \

  ln -sf /dev/stdout /var/log/nginx/access.log          &&  \
  ln -sf /dev/stderr /var/log/nginx/error.log


ADD nginx.conf /etc/nginx/nginx.conf
ADD ./entrypoint.sh /usr/local/bin/entrypoint
RUN chmod 755 /usr/local/bin/entrypoint

ENTRYPOINT [ "/usr/local/bin/entrypoint" ]
CMD [ "-start" ]

COPY srv /var/lib/initial/idpbuilder-resources
