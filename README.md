# Sample usage

## Dry run option
``` shell
./recover-delete -endpoint localhost:9000 -ak minio -sk minio123 -bucket mybucket -prefix "path/to/prefix" -dry-run
```

## To actually recover deleted objects
``` shell
./recover-delete -endpoint localhost:9000 -ak minio -sk minio123 -bucket mybucket -prefix "path/to/prefix"
```
