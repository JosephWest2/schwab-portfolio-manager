package schwabTypes

import "time"

type AllAccountsResponse []struct {
	SecuritiesAccount SecuritiesAccount `json:"securitiesAccount"`
}

type AccountResponse struct {
	SecuritiesAccount SecuritiesAccount `json:"securitiesAccount"`
}

type SecuritiesAccount struct {
	AccountNumber           string            `json:"accountNumber"`
	RoundTrips              int               `json:"roundTrips"`
	IsDayTrader             bool              `json:"isDayTrader"`
	IsClosingOnlyRestricted bool              `json:"isClosingOnlyRestricted"`
	PfcbFlag                bool              `json:"pfcbFlag"`
	Positions               []Position        `json:"positions"`
	InitialBalances         BalancesInitial   `json:"initialBalances"`
	CurrentBalances         BalancesCurrent   `json:"currentBalances"`
	ProjectedBalances       BalancesProjected `json:"projectedBalances"`
}

type Position struct {
	ShortQuantity                float64    `json:"shortQuantity"`
	AveragePrice                 float64    `json:"averagePrice"`
	CurrentDayProfitLoss         float64    `json:"currentDayProfitLoss"`
	CurrentDayProfitLossPct      float64    `json:"currentDayProfitLossPercentage"`
	LongQuantity                 float64    `json:"longQuantity"`
	SettledLongQuantity          float64    `json:"settledLongQuantity"`
	SettledShortQuantity         float64    `json:"settledShortQuantity"`
	AgedQuantity                 float64    `json:"agedQuantity"`
	Instrument                   Instrument `json:"instrument"`
	MarketValue                  float64    `json:"marketValue"`
	MaintenanceRequirement       float64    `json:"maintenanceRequirement"`
	AverageLongPrice             float64    `json:"averageLongPrice"`
	AverageShortPrice            float64    `json:"averageShortPrice"`
	TaxLotAverageLongPrice       float64    `json:"taxLotAverageLongPrice"`
	TaxLotAverageShortPrice      float64    `json:"taxLotAverageShortPrice"`
	LongOpenProfitLoss           float64    `json:"longOpenProfitLoss"`
	ShortOpenProfitLoss          float64    `json:"shortOpenProfitLoss"`
	PreviousSessionLongQuantity  float64    `json:"previousSessionLongQuantity"`
	PreviousSessionShortQuantity float64    `json:"previousSessionShortQuantity"`
	CurrentDayCost               float64    `json:"currentDayCost"`
}

type Instrument struct {
	Cusip        string  `json:"cusip"`
	Symbol       string  `json:"symbol"`
	Description  string  `json:"description"`
	InstrumentID int     `json:"instrumentId"`
	NetChange    float64 `json:"netChange"`
	Type         string  `json:"type"`
}

type BalancesInitial struct {
	AccruedInterest                  float64 `json:"accruedInterest"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade"`
	BondValue                        float64 `json:"bondValue"`
	BuyingPower                      float64 `json:"buyingPower"`
	CashBalance                      float64 `json:"cashBalance"`
	CashAvailableForTrading          float64 `json:"cashAvailableForTrading"`
	CashReceipts                     float64 `json:"cashReceipts"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower"`
	DayTradingBuyingPowerCall        float64 `json:"dayTradingBuyingPowerCall"`
	DayTradingEquityCall             float64 `json:"dayTradingEquityCall"`
	Equity                           float64 `json:"equity"`
	EquityPercentage                 float64 `json:"equityPercentage"`
	LiquidationValue                 float64 `json:"liquidationValue"`
	LongMarginValue                  float64 `json:"longMarginValue"`
	LongOptionMarketValue            float64 `json:"longOptionMarketValue"`
	LongStockValue                   float64 `json:"longStockValue"`
	MaintenanceCall                  float64 `json:"maintenanceCall"`
	MaintenanceRequirement           float64 `json:"maintenanceRequirement"`
	Margin                           float64 `json:"margin"`
	MarginEquity                     float64 `json:"marginEquity"`
	MoneyMarketFund                  float64 `json:"moneyMarketFund"`
	MutualFundValue                  float64 `json:"mutualFundValue"`
	RegTCall                         float64 `json:"regTCall"`
	ShortMarginValue                 float64 `json:"shortMarginValue"`
	ShortOptionMarketValue           float64 `json:"shortOptionMarketValue"`
	ShortStockValue                  float64 `json:"shortStockValue"`
	TotalCash                        float64 `json:"totalCash"`
	IsInCall                         bool    `json:"isInCall"`
	UnsettledCash                    float64 `json:"unsettledCash"`
	PendingDeposits                  float64 `json:"pendingDeposits"`
	MarginBalance                    float64 `json:"marginBalance"`
	ShortBalance                     float64 `json:"shortBalance"`
	AccountValue                     float64 `json:"accountValue"`
}

type BalancesCurrent struct {
	AvailableFunds                   float64 `json:"availableFunds"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade"`
	BuyingPower                      float64 `json:"buyingPower"`
	BuyingPowerNonMarginableTrade    float64 `json:"buyingPowerNonMarginableTrade"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower"`
	DayTradingBuyingPowerCall        float64 `json:"dayTradingBuyingPowerCall"`
	Equity                           float64 `json:"equity"`
	EquityPercentage                 float64 `json:"equityPercentage"`
	LongMarginValue                  float64 `json:"longMarginValue"`
	MaintenanceCall                  float64 `json:"maintenanceCall"`
	MaintenanceRequirement           float64 `json:"maintenanceRequirement"`
	MarginBalance                    float64 `json:"marginBalance"`
	RegTCall                         float64 `json:"regTCall"`
	ShortBalance                     float64 `json:"shortBalance"`
	ShortMarginValue                 float64 `json:"shortMarginValue"`
	Sma                              float64 `json:"sma"`
	IsInCall                         float64 `json:"isInCall"`
	StockBuyingPower                 float64 `json:"stockBuyingPower"`
	OptionBuyingPower                float64 `json:"optionBuyingPower"`
}

type BalancesProjected = BalancesCurrent // Same structure

type AccountNumbers struct {
	AccountNumber string `json:"accountNumber"`
	HashValue     string `json:"hashValue"`
}

type AccountNumbersResponse []AccountNumbers

type Order struct {
	Session                  string          `json:"session"`
	Duration                 string          `json:"duration"`
	OrderType                string          `json:"orderType"`
	CancelTime               time.Time       `json:"cancelTime"`
	ComplexOrderStrategyType string          `json:"complexOrderStrategyType"`
	Quantity                 float64         `json:"quantity"`
	FilledQuantity           float64         `json:"filledQuantity"`
	RemainingQuantity        float64         `json:"remainingQuantity"`
	DestinationLinkName      string          `json:"destinationLinkName"`
	ReleaseTime              time.Time       `json:"releaseTime"`
	StopPrice                float64         `json:"stopPrice"`
	StopPriceLinkBasis       string          `json:"stopPriceLinkBasis"`
	StopPriceLinkType        string          `json:"stopPriceLinkType"`
	StopPriceOffset          float64         `json:"stopPriceOffset"`
	StopType                 string          `json:"stopType"`
	PriceLinkBasis           string          `json:"priceLinkBasis"`
	PriceLinkType            string          `json:"priceLinkType"`
	Price                    float64         `json:"price"`
	TaxLotMethod             string          `json:"taxLotMethod"`
	OrderLegCollection       []OrderLeg      `json:"orderLegCollection"`
	ActivationPrice          float64         `json:"activationPrice"`
	SpecialInstruction       string          `json:"specialInstruction"`
	OrderStrategyType        string          `json:"orderStrategyType"`
	OrderID                  int64           `json:"orderId"`
	Cancelable               bool            `json:"cancelable"`
	Editable                 bool            `json:"editable"`
	Status                   string          `json:"status"`
	EnteredTime              time.Time       `json:"enteredTime"`
	CloseTime                time.Time       `json:"closeTime"`
	AccountNumber            int64           `json:"accountNumber"`
	OrderActivityCollection  []OrderActivity `json:"orderActivityCollection"`
	ReplacingOrderCollection []string        `json:"replacingOrderCollection"`
	ChildOrderStrategies     []string        `json:"childOrderStrategies"`
	StatusDescription        string          `json:"statusDescription"`
}

type OrderLeg struct {
	OrderLegType   string     `json:"orderLegType"`
	LegID          int64      `json:"legId"`
	Instrument     Instrument `json:"instrument"`
	Instruction    string     `json:"instruction"`
	PositionEffect string     `json:"positionEffect"`
	Quantity       float64    `json:"quantity"`
	QuantityType   string     `json:"quantityType"`
	DivCapGains    string     `json:"divCapGains"`
	ToSymbol       string     `json:"toSymbol"`
}

type OrderActivity struct {
	ActivityType           string         `json:"activityType"`
	ExecutionType          string         `json:"executionType"`
	Quantity               float64        `json:"quantity"`
	OrderRemainingQuantity float64        `json:"orderRemainingQuantity"`
	ExecutionLegs          []ExecutionLeg `json:"executionLegs"`
}

type ExecutionLeg struct {
	LegID             int64     `json:"legId"`
	Price             float64   `json:"price"`
	Quantity          float64   `json:"quantity"`
	MismarkedQuantity float64   `json:"mismarkedQuantity"`
	InstrumentID      int64     `json:"instrumentId"`
	Time              time.Time `json:"time"`
}
