# CSC-Staking

A staking tool for [Coinex Smart Chain](https://www.coinex.org/), used to staking CET to validators. 
It uses `Wallet Connect` protocol to connect to your wallet without providing secret key.

### Build

- Run `build_bridge.sh` to build `walletconnect-bridge`.
- Run `go build` to build the repo.

### Usage

Keep your phone and computer connected to a same LAN.

```
Usage of ./csc:
  -validator string
        validator address (default "0x62f7f2f03dc042baf765003ff0f4011720a20596")
```

Then scan displayed QR code with your wallet APP, and approve the transaction.

If you stake to the default validator "0x62f7f2f03dc042baf765003ff0f4011720a20596", you can enjoy rewards dividends. See [http://t.hk.uy/5jk](http://t.hk.uy/5jk).
