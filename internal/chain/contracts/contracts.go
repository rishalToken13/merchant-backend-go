package contracts

import (
	"encoding/json"
	"fmt"
	"os"
)

type TronContract struct {
	Address string          `json:"address"`
	ABI     json.RawMessage `json:"abi"`
}

type ContractJSON struct {
	Network string `json:"network"`

	MerchantRegistryV1_ADDRESS_TRON string          `json:"MerchantRegistryV1_ADDRESS_TRON"`
	MerchantRegistryV1_ABI          json.RawMessage `json:"MerchantRegistryV1_ABI"`

	PaymentCoreV1_ADDRESS_TRON string          `json:"PaymentCoreV1_ADDRESS_TRON,omitempty"`
	PaymentCoreV1_ABI          json.RawMessage `json:"PaymentCoreV1_ABI,omitempty"`

	TRC_20_USDT_ADDRESS_TRON string          `json:"TRC_20_USDT_ADDRESS_TRON,omitempty"`
	TRC_20_ABI               json.RawMessage `json:"TRC_20_ABI,omitempty"`
}

type Bundle struct {
	Network          string
	MerchantRegistry TronContract
	PaymentCore      TronContract
	USDT             TronContract
}

func Load(path string) (*Bundle, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read contracts json: %w", err)
	}

	var c ContractJSON
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse contracts json: %w", err)
	}

	out := &Bundle{
		Network: c.Network,
		MerchantRegistry: TronContract{
			Address: c.MerchantRegistryV1_ADDRESS_TRON,
			ABI:     c.MerchantRegistryV1_ABI,
		},
		PaymentCore: TronContract{
			Address: c.PaymentCoreV1_ADDRESS_TRON,
			ABI:     c.PaymentCoreV1_ABI,
		},
		USDT: TronContract{
			Address: c.TRC_20_USDT_ADDRESS_TRON,
			ABI:     c.TRC_20_ABI,
		},
	}

	if out.MerchantRegistry.Address == "" || len(out.MerchantRegistry.ABI) == 0 {
		return nil, fmt.Errorf("MerchantRegistryV1 address/abi missing in %s", path)
	}

	return out, nil
}
