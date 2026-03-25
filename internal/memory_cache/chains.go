package memorycache

import "deposit-collector/internal/system"

/*
ChainsCache is a cache of the supported chains and tokens.
This is a temporal solution to avoid hitting the database for each request.
In the future, we will use other caching solutions like Redis.
*/

type ChainsCache struct {
	supportedChainsByChainName map[string]*system.SupportedChain
	supportedChainsByPlatform  map[system.ChainPlatform][]*system.SupportedChain
	tokensByChainName          map[string][]*system.TokenAddress
}

func (c *ChainsCache) GetSupportedChainsByChainName(
	chainName string,
) *system.SupportedChain {
	return c.supportedChainsByChainName[chainName]
}

func (c *ChainsCache) GetSupportedChainsByPlatform(
	platform system.ChainPlatform,
) []*system.SupportedChain {
	return c.supportedChainsByPlatform[platform]
}

func (c *ChainsCache) GetPlatformByChainName(
	chainName string,
) system.ChainPlatform {
	return c.supportedChainsByChainName[chainName].ChainPlatform
}

func (c *ChainsCache) GetTokenAddressByChainNameAndTokenAddress(
	chainName string,
	tokenAddress string,
) *system.TokenAddress {
	for _, token := range c.tokensByChainName[chainName] {
		if token.Address == tokenAddress {
			return token
		}
	}
	return nil
}

func (c *ChainsCache) GetTokensByChainName(
	chainName string,
) []*system.TokenAddress {
	return c.tokensByChainName[chainName]
}

func (c *ChainsCache) GetTokenByChainNameAndTokenAddress(
	chainName string,
	tokenAddress string,
) *system.TokenAddress {
	for _, token := range c.tokensByChainName[chainName] {
		if token.Address == tokenAddress {
			return token
		}
	}
	return nil
}

func NewChainsCache(
	systemRepository *system.SystemRepository,
) (*ChainsCache, error) {
	supportedChains, err := systemRepository.GetSupportedChains()

	if err != nil {
		return nil, err
	}

	tokensByChainName := make(map[string][]*system.TokenAddress)
	for {
		tokenAddresses, err := systemRepository.GetTokenAddresses(
			system.GetTokenAddressesRequest{
				Chain:      nil,
				Address:    nil,
				UnitSymbol: nil,
				Limit:      100,
				Offset:     0,
			},
		)
		if err != nil {
			return nil, err
		}
		for _, token := range tokenAddresses {
			tokensByChainName[token.Chain.ChainName] = append(
				tokensByChainName[token.Chain.ChainName], &token,
			)
		}
		if len(tokenAddresses) < 100 {
			break
		}
	}

	supportedChainsByChainName := make(map[string]*system.SupportedChain)
	supportedChainsByPlatform := make(
		map[system.ChainPlatform][]*system.SupportedChain,
	)

	for _, chain := range supportedChains {
		supportedChainsByChainName[chain.ChainName] = &chain
		supportedChainsByPlatform[chain.ChainPlatform] = append(
			supportedChainsByPlatform[chain.ChainPlatform], &chain,
		)
	}

	return &ChainsCache{
		supportedChainsByChainName: supportedChainsByChainName,
		supportedChainsByPlatform:  supportedChainsByPlatform,
		tokensByChainName:          tokensByChainName,
	}, nil
}
