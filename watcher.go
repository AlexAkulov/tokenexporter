package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type watchItem struct {
	Token    string
	Wallet   string
	Decimals int

	Metric prometheus.Gauge
}

type Watcher struct {
	geth  map[string]*ethclient.Client
	items map[string][]*watchItem
}

func (w Watcher) addGeth(name, url string) error {
	eth, err := ethclient.Dial(url)
	if err != nil {
		return err
	}
	w.geth[name] = eth
	return nil
}

func (w *Watcher) LoadConfig(config *ConfigYAML) error {
	for name, url := range config.Chains {
		if err := w.addGeth(name, url); err != nil {
			return err
		}
	}
	tokensMap := map[string]map[string]Token{}
	for chain, tokens := range config.Tokens {
		tokensMapChain := map[string]Token{}
		for _, token := range tokens {
			tokensMapChain[token.Symbol] = token
		}
		tokensMap[chain] = tokensMapChain
	}
	for _, wallet := range config.Wallets {
		for chain, symbols := range wallet.TrackFor {
			if _, ok := config.Chains[chain]; !ok {
				return fmt.Errorf("Unknown chain '%s' in wallet '%s'", chain, wallet.Address)
			}
			for _, symbol := range symbols {
				token, ok := tokensMap[chain][symbol]
				if !ok {
					return fmt.Errorf("Unknown token '%s' for chain '%s' in wallet '%s'", symbol, chain, wallet.Address)
				}
				labels := map[string]string{}
				for k, v := range wallet.Labels {
					labels[k] = v
				}
				labels["symbol"] = symbol
				labels["name"] = wallet.Name
				labels["wallet"] = wallet.Address
				labels["token"] = token.Contract
				labels["chain"] = chain

				metric := prometheus.NewGauge(prometheus.GaugeOpts{
					Name:        "token_balance",
					ConstLabels: labels,
				})
				item := watchItem{
					Token:    token.Contract,
					Wallet:   wallet.Address,
					Decimals: token.Decimal,

					Metric: metric,
				}
				w.items[chain] = append(w.items[chain], &item)
			}
		}
	}
	return nil
}

func (w Watcher) Start(listen string) error {
	registry := prometheus.NewRegistry()

	for chain := range w.items {
		for _, item := range w.items[chain] {
			if err := registry.Register(item.Metric); err != nil {
				return err
			}
		}
	}
	http.Handle(
		"/metrics", promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: false,
			}),
	)
	// To test: curl -H 'Accept: application/openmetrics-text' localhost:8080/metrics
	go func() {
		err := http.ListenAndServe(listen, nil)
		if err != nil {
			log.Println("can't start listen:", err)
			os.Exit(1)
		}
	}()

	go func() {
		for {
			for chain := range w.items {
				go func(c string) {
					log.Printf("Updating tokens on '%s'", c)
					for _, item := range w.items[c] {
						balance := w.getTokenBalance(c, item)
						item.Metric.Set(balance)
					}
					log.Printf("Updated tokens on '%s'", c)
				}(chain)
			}
			time.Sleep(10 * time.Minute)
		}
	}()
	return nil
}

func (w Watcher) Stop() error {
	return nil
}

func (w *Watcher) tokenCaller(eth *ethclient.Client, address common.Address) (*MainCaller, error) {
	caller, err := NewMainCaller(address, eth)
	if err != nil {
		return nil, err
	}
	return caller, err
}

// Fetch ETH balance from Geth server
func (w *Watcher) getTokenBalance(chain string, item *watchItem) float64 {
	if (item.Token == "0x0000000000000000000000000000000000000000") ||
		(item.Token == "") {
		return w.getGasBalance(chain, item)
	}

	caller, err := w.tokenCaller(w.geth[chain], common.HexToAddress(item.Token))
	if err != nil {
		log.Printf("Can't tokenCaller for chain='%s', wallet='%s', token='%s': %v", chain, item.Wallet, item.Token, err)
		return -1
	}
	balance, err := caller.BalanceOf(nil, common.HexToAddress(item.Wallet))
	if err != nil {
		log.Printf("Can't BalanceOf for chain='%s', wallet='%s', token='%s': %v", chain, item.Wallet, item.Token, err)
		return -1
	}
	return toFloat(balance, item.Decimals)
}

func (w *Watcher) getGasBalance(chain string, item *watchItem) float64 {
	balance, err := w.geth[chain].BalanceAt(
		context.Background(), common.HexToAddress(item.Wallet), nil,
	)
	if err != nil {
		log.Printf("Can't BalanceAt for chain='%s', wallet='%s', token='%s': %v", chain, item.Wallet, item.Token, err)
		return -1
	}
	return toFloat(balance, item.Decimals)
}

func toFloat(balanceAt *big.Int, decimails int) float64 {
	fbalance := new(big.Float)
	fbalance.SetString(balanceAt.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(decimails)))
	result, _ := ethValue.Float64()
	return result
}
