# Conditional writes

"_The lost update problem_" is usually fixed with conditional writes using
`ETag` and the `If-Match` header. This is a form of optimistic locking to ensure
that the user doesn't overwrite previous changes.

> TLDR;
>
> In the context of PATCH requests on RLS-enabled resources, facing the need to
> do conditional writes we do read-and-update in a transaction
> with REPEATABLE READ isolation level, over the more common `and etag=<etag>`
> pattern, to achieve a clear distinction between lost updates and lacking
> access, accepting that we need to do an additional read.

## How it works

This is how it works conceptually. Comments are handled by the application code.

```sql
BEGIN READ WRITE ISOLATION LEVEL REPEATABLE READ;
-- Set role and user info for RLS

SELECT a, etag from resource where id = 1;
-- Dependent on the result, do one of
--  * 0 rows returned - 404 Not found or not authorized
--  * Compare etag and return 412 Precondition failed if none match.
--  * Compare resource and return 200 if the update won't do anything.

UPDATE resource set a = 'upd' where id = 1 RETURNING *;
-- Dependent on the result, do one of the following
-- If ERROR: could not serialize access due to concurrent update - 412 Precondition failed
-- If affected rows = 0 - 403 Forbidden (since we have confirmed the users read access above)
-- If affected rows = 1 - 200 OK

COMMIT;
