# .hl Resolver

This is the original (hackish) resolver server used by hl.baby and hl.place. Currently, we use the demo API key because Hyperliquid Names were still in its early stage.

There are some notably acceptable TODOs on code readability and conventions (a good first issue.)

## Building
```sh
# must have golang version ^1.24.1
make
```

## Requirements
We try to make the program as efficient as possible so it can run anywhere. This is the minimum server specs:

```
3000MHz CPU with 1 logical threads
2GB RAM
10GB SSD
1Gbit/s network uplink
Debian 12
```

Aim for higher uplink, storage, and memory it's the main priority.