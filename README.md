# Skribserv

Servilo por io. Bedaŭrinde, mi devas programi angle, ĉar miksi du lingvojn ene un unuopaj vortoj dolorigas mian cerbon.

## Datumbazo

```sh
podman run -it --rm -p 5432:5432 -e "POSTGRES_USER=skribserv" -e "POSTGRES_PASSWORD=skribserv" -e "POSTGRES_DB=skribserv" postgres
psql "postgres://skribserv:skribserv@127.0.0.1:5432/skribserv"
INSERT INTO users (id, name, email, password, admin, created_at, updated_at) VALUES ('admin', 'admin', 'admin', 'admin', true, NOW(), NOW()) ;
```