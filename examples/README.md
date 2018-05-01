# bitmark-sdk-go example

[Testnet Registry](https://registry.test.bitmark.com)
[API Doc](https://bitmarkcoreapi.docs.apiary.io)

### Create New Account

```bash
go run main.go new-account
```

### Issue

[Bitmark Account](https://registry.test.bitmark.com/account/efPvC7aGuZL4Hx19nEJZ4shmmGdAUXYeDE6peQdQ6HT6Ua3oao/owned)

```bash
echo `date +%s` > test.txt
go run main.go issue \
    -issuer 5XEECtre5nDKsLSzL4jPuPyPJ5jccf1EPN6cZWNnh8pnxjyicscmE1n \
    -acs "public" \
    -p test.txt \
    -name "Test issue file" \
    -meta "KEY:value"
```

```bash
echo `date +%s` > test.txt
go run main.go issue \
    -issuer 5XEECtre5nDKsLSzL4jPuPyPJ5jccf1EPN6cZWNnh8pnxjyicscmE1n \
    -acs "private" \
    -p test.txt \
    -name "Test file" \
    -meta "KEY:value"
```

### Issue by asset Id

```bash
go run main.go issue-asset-id -aid b0f3e9ebca129d13c7986aa384b29c169eaa3f9406c02689216c72af83246df861b8e07bbbf250b1b24d3397f38c870fcdaff3a278c6a3f86f12504c7ea24164 -issuer 5XEECtre5nDKsLSzL4jPuPyPJ5jccf1EPN6cZWNnh8pnxjyicscmE1n
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