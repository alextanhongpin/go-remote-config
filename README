# go-remote-config

Attempt for real-time remote config with postgres Listen/Notify and golang's sync.Map that is concurrent-safe.


To update remote config, we can just use this.
```sql
select pg_notify('config', '{"name": "john", "age": 10}');

notify config, '{"name": "john", "age": 10}';
```

A better way is to store each key-value in a table. Then attach a trigger that will call pg notify. This improves visibility compared to keeping the key-value in redis.



