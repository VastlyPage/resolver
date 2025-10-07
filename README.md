# .hl Resolver

.hl content resolver that supports mirroring regular web2 content and IPFS content.

## Setup
You can run your own resolver locally using this instruction:
```sh
git clone https://github.com/VastlyPage/resolver
docker compose up -d
docker inspect resolver-resolver-1 | grep IPAddress
curl -k https://ip_address -H 'Host: docs.hl.place' -v
```