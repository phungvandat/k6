K6 testing

Start HTTP server:
- init env (first time)
    ```
    cat .env.example > .env
    ```
- run server:
    ```
    make dev
    ```

K6 test APIs:
```bash
    # get API
    make test_get

    # post API
    make test_post

    # post sync
    make test_post_sync

    # post async
    make test_post_async

    # post batch
    make test_post_batch
``` 