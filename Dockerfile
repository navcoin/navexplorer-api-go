FROM iron/base

EXPOSE 8888
ADD navexplorer-api-linux-amd64 /
ENTRYPOINT ["./navexplorer-api-linux-amd64"]