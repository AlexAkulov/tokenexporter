# Token Exporter

A lightweight Prometheus exporter that will output ERC20 Token balances from a list of addresses you specify. Token Exporter attaches to a geth servers to fetch token wallet balances for your Grafana dashboards and alerts.


## Build
### Generate erc20.go
https://goethereumbook.org/en/smart-contract-read-erc20
```
abigen --abi=abi --pkg=main --out=erc20.go
```

## Configuration

```
chains:
  fantom: https://rpc.ftm.tools
  bsc: https://bsc-dataseed1.binance.org
  polygon: https://polygon-rpc.com

tokens:
  fantom:
    - symbol: FTM            # Gas
    - symbol: USDC
      contract: "0x04068DA6C83AFCFA0e13ba15A6696662335D5B75"
      decimal: 18            # default 18
    - symbol: fUSDT
      contract: "0x049d68029688eAbF473097a2fC38ef61633A3C7A"
  bsc:
    - symbol: BNB
    - symbol: USDC
      contract: "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d"
  polygon:
    - symbol: MATIC

wallets:
  - name: KillerWhale0063
    address: "0x258473a955e900385f984fb3fbfd7480d5949cd7"
    track_for:
      fantom:
        - FTM
        - USDC
      bsc:
        - BNB
        - USDC
    labels:
      type: protocol
```

## Run

```
LISTEN=":9015" CONFIG_FILE="config.yml" ./tokenexporter`
```

## Output

```
curl "http://localhost:9015/metrics"
# HELP token_balance
# TYPE token_balance gauge
token_balance{chain="bsc",name="KillerWhale0063",symbol="BNB",token="",type="protocol",wallet="0x258473a955e900385f984fb3fbfd7480d5949cd7"} 80.0815651
token_balance{chain="bsc",name="KillerWhale0063",symbol="USDC",token="0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d",type="protocol",wallet="0x258473a955e900385f984fb3fbfd7480d5949cd7"} 0
token_balance{chain="fantom",name="KillerWhale0063",symbol="FTM",token="",type="protocol",wallet="0x258473a955e900385f984fb3fbfd7480d5949cd7"} 9.473403423496291e+06
token_balance{chain="fantom",name="KillerWhale0063",symbol="USDC",token="0x04068DA6C83AFCFA0e13ba15A6696662335D5B75",type="protocol",wallet="0x258473a955e900385f984fb3fbfd7480d5949cd7"} 1.5064170747e-08
```
