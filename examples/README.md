# bitmark-sdk-go example

### Create New Account

```bash
go run main.go new-account
```

### Issue by asset

[Bitmark Account](https://registry.test.bitmark.com/account/efPvC7aGuZL4Hx19nEJZ4shmmGdAUXYeDE6peQdQ6HT6Ua3oao/owned)

```bash
go run main.go issue-asset-file \
    -issuer 5XEECtre5nDKsLSzL4jPuPyPJ5jccf1EPN6cZWNnh8pnxjyicscmE1n \
    -p test.jpg \
    -name "晴天娃娃" \
    -meta="DESCRIPTION:晴天娃娃照片"
```

### Download asset

```bash
go run main.go download \
    -owner 5XEECtgwKTikNY17b7NWjYz5LD39tJzhoEThW3oZ8vZ8rdngkdqcGY7 \
    -bid ba9d57354f0a1847be3e6c3f7e96068c2015b8503cbdeaecaf553ff776371aea
```

### Querying a specific bitmark

[PROPERTY BITMARK](https://registry.test.bitmark.com/bitmark/ba9d57354f0a1847be3e6c3f7e96068c2015b8503cbdeaecaf553ff776371aea)

```bash
go run main.go bitmark \
    -bid ba9d57354f0a1847be3e6c3f7e96068c2015b8503cbdeaecaf553ff776371aea \
    -issuer 5XEECtre5nDKsLSzL4jPuPyPJ5jccf1EPN6cZWNnh8pnxjyicscmE1n
```

### Querying a set of bitmarks

```bash
go run main.go bitmarks \
    -issuer 5XEECtre5nDKsLSzL4jPuPyPJ5jccf1EPN6cZWNnh8pnxjyicscmE1n
```