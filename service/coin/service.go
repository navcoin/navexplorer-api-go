package coin

import "github.com/NavExplorer/navexplorer-api-go/config"

var IndexAddressTransaction = config.Get().SelectedNetwork + ".addresstransaction"

func GetWealthDistribution(groups []int) (err error) {
	return
}