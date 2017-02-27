FROM golang:onbuild

ENV MARTINI_ENV production

VOLUME /etc/steam-item-finder

EXPOSE 3101
