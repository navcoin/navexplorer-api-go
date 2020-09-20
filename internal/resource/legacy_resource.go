package resource

import (
	"errors"
	"fmt"
	"github.com/NavExplorer/navexplorer-api-go/internal/framework/pagination"
	"github.com/NavExplorer/navexplorer-api-go/internal/repository"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/address"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/block"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/coin"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/dao"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/group"
	"github.com/NavExplorer/navexplorer-api-go/internal/service/softfork"
	"github.com/NavExplorer/navexplorer-indexer-go/pkg/explorer"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type LegacyResource struct {
	addressService  address.Service
	blockService    block.Service
	coinService     coin.Service
	daoService      dao.Service
	softForkService softfork.Service
}

func NewLegacyResource(
	addressService address.Service,
	blockService block.Service,
	coinService coin.Service,
	daoService dao.Service,
	softForkService softfork.Service,
) *LegacyResource {
	return &LegacyResource{
		addressService,
		blockService,
		coinService,
		daoService,
		softForkService,
	}
}

// Address Resources
func (r *LegacyResource) GetAddress(c *gin.Context) {
	a, err := r.addressService.GetAddress(c.Param("hash"))
	if err != nil {
		if err == repository.ErrAddressNotFound {
			handleError(c, err, http.StatusNotFound)
		} else if err == repository.ErrAddressInvalid {
			handleError(c, err, http.StatusBadRequest)
		} else {
			handleError(c, err, http.StatusInternalServerError)
		}
		return
	}

	c.JSON(200, a)
}

func (r *LegacyResource) GetAddresses(c *gin.Context) {
	config, _ := pagination.Bind(c)

	a, total, err := r.addressService.GetAddresses(config)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	paginator := pagination.NewPaginator(len(a), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, a)
}

func (r *LegacyResource) GetBalanceChart(c *gin.Context) {
	chart, err := r.addressService.GetBalanceChart(c.Param("hash"))
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, chart)
}

func (r *LegacyResource) GetStakingChart(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")

	chart, err := r.addressService.GetStakingChart(period, c.Param("hash"))
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, chart)
}

func (r *LegacyResource) GetAssociatedStakingAddresses(c *gin.Context) {
	addresses, err := r.addressService.GetAssociatedStakingAddresses(c.Param("hash"))
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, addresses)
}

func (r *LegacyResource) GetBalancesForAddresses(c *gin.Context) {
	_ = c.Request.ParseForm()

	addresses := make([]string, 0)
	if addressesParam := c.Request.Form.Get("addresses"); addressesParam != "" {
		addresses = strings.Split(addressesParam, ",")
	}

	balances, err := r.addressService.GetNamedAddresses(addresses)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, balances)
}

func (r *LegacyResource) GetBestBlock(c *gin.Context) {
	b, err := r.blockService.GetBestBlock()
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, b.Height)
}

func (r *LegacyResource) GetBlockGroups(c *gin.Context) {
	period := group.GetPeriod(c.DefaultQuery("period", "daily"))
	if period == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("Invalid period `%s`", c.Query("period")),
			"status":  http.StatusBadRequest,
		})
		return
	}

	count, err := strconv.Atoi(c.DefaultQuery("count", "10"))
	if err != nil || count < 10 {
		count = 10
	}

	groups, err := r.blockService.GetBlockGroups(period, count)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, groups)
}

func (r *LegacyResource) GetBlocks(c *gin.Context) {
	config, _ := pagination.Bind(c)

	blocks, total, err := r.blockService.GetBlocks(config)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	paginator := pagination.NewPaginator(len(blocks), total, config)
	paginator.WriteHeader(c)

	c.JSON(200, blocks)
}

func (r *LegacyResource) GetBlock(c *gin.Context) {
	b, err := r.blockService.GetBlock(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		handleError(c, err, http.StatusInternalServerError)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, b)
}

func (r *LegacyResource) GetBlockTransactions(c *gin.Context) {
	b, err := r.blockService.GetBlock(c.Param("hash"))
	txs, err := r.blockService.GetTransactions(b.Hash)
	if err == repository.ErrBlockNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	if txs == nil {
		txs = make([]*explorer.BlockTransaction, 0)
	}

	c.JSON(200, txs)
}

func (r *LegacyResource) GetRawBlock(c *gin.Context) {
	b, err := r.blockService.GetRawBlock(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, b)
}

func (r *LegacyResource) GetTransaction(c *gin.Context) {
	tx, err := r.blockService.GetTransactionByHash(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, tx)
}

func (r *LegacyResource) GetRawTransaction(c *gin.Context) {
	tx, err := r.blockService.GetRawTransactionByHash(c.Param("hash"))
	if err == repository.ErrBlockNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, tx)
}

func (r *LegacyResource) GetWealthDistribution(c *gin.Context) {
	groupsQuery := c.DefaultQuery("groups", "10,100,1000")
	if groupsQuery == "" {
		groupsQuery = "10,100,1000"
	}

	groups := make([]string, 0)
	groups = strings.Split(groupsQuery, ",")

	b := make([]int, len(groups))
	for i, v := range groups {
		b[i], _ = strconv.Atoi(v)
	}

	distribution, err := r.coinService.GetWealthDistribution(b)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, distribution)
}

func (r *LegacyResource) GetBlockCycle(c *gin.Context) {
	b, err := r.blockService.GetBestBlock()
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	blockCycle, err := r.daoService.GetBlockCycleByBlock(b)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, blockCycle)
}

func (r *LegacyResource) GetCfundStats(c *gin.Context) {
	cfundStats, err := r.daoService.GetCfundStats()
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, cfundStats)
}

func (r *LegacyResource) GetProposal(c *gin.Context) {
	proposal, err := r.daoService.GetProposal(c.Param("hash"))
	if err == repository.ErrProposalNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, proposal)
}

func (r *LegacyResource) GetProposalVotingTrend(c *gin.Context) {
	trend, err := r.daoService.GetProposalTrend(c.Param("hash"))
	if err == repository.ErrProposalNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, trend)
}

func (r *LegacyResource) GetProposalVotes(c *gin.Context) {
	vote, err := strconv.ParseBool(c.Param("vote"))
	votes, err := r.daoService.GetProposalVotes(c.Param("hash"))
	if err != nil || votes == nil {
		handleError(c, err, http.StatusNotFound)
	}

	legacyVotes := make([]*Votes, 0)
	for _, v := range votes {
		for _, a := range v.Addresses {
			legacyVote := &Votes{Address: a.Address}
			if vote == true {
				legacyVote.Votes = int64(a.Yes)
			} else {
				legacyVote.Votes = int64(a.No)
			}
			legacyVotes = append(legacyVotes, legacyVote)
		}
	}

	c.JSON(200, votes)
}

func (r *LegacyResource) GetPaymentRequestsForProposal(c *gin.Context) {
	proposal, err := r.daoService.GetProposal(c.Param("hash"))
	if err == repository.ErrProposalNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	paymentRequests, err := r.daoService.GetPaymentRequestsForProposal(proposal)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, paymentRequests)
}

func (r *LegacyResource) GetPaymentRequestByHash(c *gin.Context) {
	paymentRequest, err := r.daoService.GetPaymentRequest(c.Param("hash"))
	if err == repository.ErrPaymentRequestNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, paymentRequest)
}

func (r *LegacyResource) GetPaymentRequestVotingTrend(c *gin.Context) {
	trend, err := r.daoService.GetPaymentRequestTrend(c.Param("hash"))
	if err == repository.ErrPaymentRequestNotFound {
		handleError(c, err, http.StatusNotFound)
		return
	}
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, trend)
}

func (r *LegacyResource) GetPaymentRequestVotes(c *gin.Context) {
	vote, err := strconv.ParseBool(c.Param("vote"))
	votes, err := r.daoService.GetPaymentRequestVotes(c.Param("hash"))
	if err != nil || votes == nil {
		handleError(c, err, http.StatusNotFound)
	}

	legacyVotes := make([]*Votes, 0)
	for _, v := range votes {
		for _, a := range v.Addresses {
			legacyVote := &Votes{Address: a.Address}
			if vote == true {
				legacyVote.Votes = int64(a.Yes)
			} else {
				legacyVote.Votes = int64(a.No)
			}
			legacyVotes = append(legacyVotes, legacyVote)
		}
	}

	c.JSON(200, votes)
}

func (r *LegacyResource) Search(c *gin.Context) {
	query := c.Query("query")

	var result Result
	var err error

	_, err = r.daoService.GetProposal(query)
	if err == nil {
		result.Type = "proposal"
		result.Value = query
		c.JSON(200, result)
		return
	}

	_, err = r.daoService.GetPaymentRequest(query)
	if err == nil {
		result.Type = "paymentRequest"
		result.Value = query
		c.JSON(200, result)
		return
	}

	_, err = r.blockService.GetBlock(query)
	if err == nil {
		result.Type = "block"
		result.Value = query
		c.JSON(200, result)
		return
	}

	_, err = r.blockService.GetTransactionByHash(query)
	if err == nil {
		result.Type = "transaction"
		result.Value = query
		c.JSON(200, result)
		return
	}

	_, err = r.addressService.GetAddress(query)
	if err == nil {
		result.Type = "address"
		result.Value = query
		c.JSON(200, result)
		return
	}

	handleError(c, errors.New("no search result"), http.StatusNotFound)
}

func (r *LegacyResource) GetSoftForks(c *gin.Context) {
	softForks, err := r.softForkService.GetSoftForks()
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, softForks)
}

func (r *LegacyResource) GetStakingReport(c *gin.Context) {
	stakingReport, err := r.addressService.GetStakingReport()
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, stakingReport)
}

func (r *LegacyResource) GetStakingByBlockCount(c *gin.Context) {
	blockCount, err := strconv.Atoi(c.DefaultQuery("blocks", "1000"))
	if err != nil {
		blockCount = 1000
	}
	if blockCount > 50000000 {
		blockCount = 50000000
	}

	staking, err := r.addressService.GetStakingByBlockCount(blockCount, false)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, staking)
}

func (r *LegacyResource) GetStakingRewardsForAddresses(c *gin.Context) {
	addresses := strings.Split(c.Query("addresses"), ",")
	if len(addresses) == 0 {
		handleError(c, errors.New("No addresses provided"), http.StatusBadRequest)
		return
	}

	rewards, err := r.addressService.GetStakingRewardsForAddresses(addresses)
	if err != nil {
		handleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, rewards)
}

type Votes struct {
	Address string `json:"address"`
	Votes   int64  `json:"votes"`
}

func handleError(c *gin.Context, err error, status int) {
	c.AbortWithStatusJSON(status, gin.H{
		"status":  status,
		"message": err.Error(),
	})
}

func urlDecodeType(txType string) string {
	txType = strings.ReplaceAll(txType, "-", "_")
	txType = strings.ToUpper(txType)

	return txType
}
