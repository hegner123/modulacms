# Page

## Datatype: Page
Label: Page

## Schema

```sql
CREATE TABLE IF NOT EXISTS datatypes
(
    datatype_id   INTEGER
        primary key,
    parent_id     INTEGER default NULL
        references datatypes
            on update cascade on delete set default,
    label         TEXT                     not null,
    type          TEXT                     not null,
    author        TEXT    default "system" not null
        references users (username)
            on update cascade on delete set default,
    author_id     INTEGER default 1        not null
        references users (user_id)
            on update cascade on delete set default,
    history TEXT,
    date_created  TEXT    default CURRENT_TIMESTAMP,
    date_modified TEXT    default CURRENT_TIMESTAMP
);
```
