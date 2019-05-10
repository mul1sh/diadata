package diaApi

import (
	"encoding/json"
	"errors"
	"github.com/diadata-org/diadata/pkg/dia"
	"github.com/diadata-org/diadata/pkg/dia/helpers"
	"github.com/diadata-org/diadata/pkg/http/restApi"
	"github.com/diadata-org/diadata/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
	"database/sql"
)

type Env struct {
	DataStore models.Datastore
}

// PostSupply godoc
// @Summary Post the circulating supply
// @Description Post the circulating supply
// @Tags dia
// @Accept  json
// @Produce  json
// @Param Symbol query string true "Coin symbol"
// @Param CirculatingSupply query float64 true "number of coins in circulating supply"
// @Success 200 {object} dia.Supply	"success"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/supply [post]
func (env *Env) PostSupply(c *gin.Context) {

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		restApi.SendError(c, http.StatusInternalServerError, errors.New("ReadAll"))
	} else {
		var t dia.Supply
		err = json.Unmarshal(body, &t)
		if err != nil {
			restApi.SendError(c, http.StatusInternalServerError, errors.New("Unmarshal"))
		} else {
			if t.Symbol == "" || t.CirculatingSupply == 0.0 {
				log.Errorln("received supply:", t)
				restApi.SendError(c, http.StatusInternalServerError, errors.New("Missing Symbol or CirculatingSupply value"))
			} else {
				log.Println("received supply:", t)
				source := dia.Diadata
				if t.Source != "" {
					source = t.Source
				}
				s := &dia.Supply{
					Time:              time.Now(),
					Name:              helpers.NameForSymbol(t.Symbol),
					Symbol:            t.Symbol,
					Source:            source,
					CirculatingSupply: t.CirculatingSupply}

				err := env.DataStore.SetSupply(s)

				if err == nil {
					c.JSON(http.StatusOK, s)
				} else {
					restApi.SendError(c, http.StatusInternalServerError, err)
				}
			}
		}
	}
}

// GetQuotation godoc
// @Summary Get quotation
// @Description GetQuotation
// @Tags dia
// @Accept  json
// @Produce  json
// @Param   symbol     path    string     true        "Some symbol"
// @Success 200 {object} models.Quotation "success"
// @Failure 404 {object} restApi.APIError "Symbol not found"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/quotation/:symbol: [get]
func (env *Env) GetQuotation(c *gin.Context) {
	symbol := c.Param("symbol")
	q, err := env.DataStore.GetQuotation(symbol)
	if err != nil {
		if err == redis.Nil {
			restApi.SendError(c, http.StatusNotFound, err)
		} else {
			restApi.SendError(c, http.StatusInternalServerError, err)
		}
	} else {
		c.JSON(http.StatusOK, q)
	}
}

// GetSupply godoc
// @Summary Get supply
// @Description GetSupply
// @Tags dia
// @Accept  json
// @Produce  json
// @Param   symbol     path    string     true        "Some symbol"
// @Success 200 {object} dia.Supply "success"
// @Failure 404 {object} restApi.APIError "Symbol not found"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/supply/:symbol: [get]
func (env *Env) GetSupply(c *gin.Context) {
	symbol := c.Param("symbol")
	s, err := env.DataStore.GetSupply(symbol)
	if err != nil {
		if err == redis.Nil {
			restApi.SendError(c, http.StatusNotFound, err)
		} else {
			restApi.SendError(c, http.StatusInternalServerError, err)
		}
	} else {
		c.JSON(http.StatusOK, s)
	}
}

// GetPairs godoc
// @Summary Get pairs
// @Description Get pairs
// @Tags dia
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Pairs "success"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/pairs/ [get]
func (env *Env) GetPairs(c *gin.Context) {
	p, err := env.DataStore.GetPairs("")
	if err != nil {
		restApi.SendError(c, http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, &models.Pairs{Pairs: p})
	}
}

// GetSymbol godoc
// @Summary Get Symbol Details
// @Description Get Symbol Details
// @Tags dia
// @Accept  json
// @Produce  json
// @Param   symbol     path    string     true        "Some symbol"
// @Success 200 {object} models.SymbolDetails "success"
// @Failure 404 {object} restApi.APIError "Symbol not found"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/symbol/:symbol: [get]
func (env *Env) GetSymbolDetails(c *gin.Context) {
	symbol := c.Param("symbol")

	s, err := env.DataStore.GetSymbolDetails(symbol)
	if err != nil {
		if err == redis.Nil {
			restApi.SendError(c, http.StatusNotFound, err)
		} else {
			restApi.SendError(c, http.StatusInternalServerError, err)
		}
	} else {
		c.JSON(http.StatusOK, s)
	}
}

func roundUpTime(t time.Time, roundOn time.Duration) time.Time {
	t = t.Round(roundOn)
	if time.Since(t) >= 0 {
		t = t.Add(roundOn)
	}
	return t
}

// GetCoins godoc
// @Summary Get coins
// @Description GetCoins
// @Tags dia
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Coins "success"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/coins [get]
func (env *Env) GetCoins(c *gin.Context) {
	coins, err := env.DataStore.GetCoins()
	if err != nil {
		restApi.SendError(c, http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, coins)
	}
}

// GetChartPoints godoc
// @Summary Get chart points for
// @Description Get Symbol Details
// @Tags dia
// @Accept  json
// @Produce  json
// @Param   symbol     path    string     true        "Some symbol"
// @Param   exchange     path    string     true        "Some exchange"
// @Param   filter     path    string     true        "Some filter"
// @Param   scale      query   string     false       "scale 5m 30m 1h 4h 1d 1w"
// @Success 200 {object} models.Points "success"
// @Failure 404 {object} restApi.APIError "Symbol not found"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/chartPoints/:filter/:exchange:/:symbol: [get]
func (env *Env) GetChartPoints(c *gin.Context) {
	filter := c.Param("filter")
	exchange := c.Param("exchange")
	symbol := c.Param("symbol")
	scale := c.Query("scale")

	p, err := env.DataStore.GetFilterPoints(filter, exchange, symbol, scale)
	if err != nil {
		restApi.SendError(c, http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, p)
	}
}

// GetChartPointsAllExchange godoc
// @Summary Get Symbol Details
// @Description Get Symbol Details
// @Tags dia
// @Accept  json
// @Produce  json
// @Param   symbol     path    string     true        "Some symbol"
// @Param   filter     path    string     true        "Some filter"
// @Param   scale      query   string     false       "scale 5m 30m 1h 4h 1d 1w"
// @Success 200 {object} models.Points "success"
// @Failure 404 {object} restApi.APIError "Symbol not found"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/chartPointsAllExchanges/:filter:/:symbol: [get]
func (env *Env) GetChartPointsAllExchanges(c *gin.Context) {
	filter := c.Param("filter")
	symbol := c.Param("symbol")
	scale := c.Query("scale")

	p, err := env.DataStore.GetFilterPoints(filter, "", symbol, scale)
	if err != nil {
		restApi.SendError(c, http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, p)
	}
}

// GetAllSymbols godoc
// @Summary Get all symbols list
// @Description Get all symbols list
// @Tags dia
// @Accept  json
// @Produce  json
// @Param   symbol     path    string     true        "Some symbol"
// @Param   filter     path    string     true        "Some filter"
// @Param   scale      query   string     false       "scale 5m 30m 1h 4h 1d 1w"
// @Success 200 {object} dia.Symbols "success"
// @Failure 500 {object} restApi.APIError "error"
// @Router /v1/symbols [get]
func (env *Env) GetAllSymbols(c *gin.Context) {
	s := env.DataStore.GetAllSymbols()
	if len(s) == 0 {
		restApi.SendError(c, http.StatusInternalServerError, errors.New("cant find symbols"))
	} else {
		c.JSON(http.StatusOK, dia.Symbols{Symbols: s})
	}
}


func (env *Env) GetAllTokenDetails(c *gin.Context){
	var (
		sto dia.Security_Token_Details
		result gin.H
	)
	db, err := sql.Open("mysql", "root:@93MySQL@/sys")
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()
	// make sure connection is available
	err = db.Ping()
	if err != nil {
		log.Print(err.Error())
	}
	token_symbol := c.Param("token_symbol")
	row := db.QueryRow("select token_name, token_status, token_symbol, industry, amount_raised, currency, issuance_price,min_invest, closing_date, target_investor_type, jurisdictions_avail, restricted_area, secondary_market, website, whitepaper, prospectus, smart_contract, github, blockchain, issuer_address, token_used, dividend, voting, equity_ownership, mme_class, interest, portfolio from SecurityTokens where token_symbol = ?;",token_symbol)

	err = row.Scan(&sto.Token_Name, &sto.Token_Status, &sto.Token_Symbol, &sto.Industry, &sto.Amount_Raised, &sto.Currency, &sto.Issuance_Price, &sto.Min_Invest, &sto.Closing_Date, &sto.Target_Investor_Type, &sto.Jurisdictions_Avail, &sto.Restricted_Area, &sto.Secondary_Market, &sto.Website, &sto.Whitepaper, &sto.Prospectus, &sto.Smart_Contract, &sto.Github, &sto.Blockchain, &sto.Issuer_Address, &sto.Token_Used, &sto.Dividend, &sto.Voting, &sto.Equity_Ownership, &sto.MME_Class, &sto.Interest, &sto.Portfolio)

	if err != nil {
		// If no results send null
		result = gin.H{
			"result": nil,
			"count":  0,
		}
	} else {
		result = gin.H{
			"result": sto,
			"count":  1,
		}
	}
	c.JSON(http.StatusOK, result)
}

func (env *Env) GetAllTokens(c *gin.Context){
	var (
		sto  dia.Security_Token_Symbols
		tokens []dia.Security_Token_Symbols
	)
	db, err := sql.Open("mysql", "root:@93MySQL@/sys")
	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()
	// make sure connection is available
	err = db.Ping()
	if err != nil {
		log.Print(err.Error())
	}
	rows, err := db.Query("select token_name, token_symbol from SecurityTokens;")
	if err != nil {
		log.Print(err.Error())
	}
	for rows.Next() {
		err = rows.Scan(&sto.Token_Name, &sto.Token_Symbol)
		tokens = append(tokens, sto)
		if err != nil {
			log.Print(err.Error())
		}
	}
	defer rows.Close()
	c.JSON(http.StatusOK, gin.H{
		"result": tokens,
		"count":  len(tokens),
	})
}
