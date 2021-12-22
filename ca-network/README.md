# Busy CA server 

Start Fabric CA server for Busy 
```
docker-compose up -d ca-database
docker-compose up -d fabric-ca-server
```

To confirm if everything is working
```
curl -k https://localhost:7054/api/v1/cainfo
```