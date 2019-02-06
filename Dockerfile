FROM scratch

EXPOSE 8080
ADD config.prod.yaml /
ADD navexplorerApi /

ENTRYPOINT ["./navexplorerApi"]