# Tests

## Catsplash Multicliente

```bash
sudo ./test_catsplash_multiclient.sh up -n 50
# ---> en OTRA terminal, recién ahora:
sudo ip netns exec ns_router ./bin/catsplash
# ---> de vuelta en la primera terminal:
sudo ./test_catsplash_multiclient.sh run
sudo ./test_catsplash_multiclient.sh down
```
