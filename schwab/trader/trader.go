package trader

import "time"

type AllAccountsResponse []struct {
	SecuritiesAccount SecuritiesAccount `json:"securitiesAccount"`
}

type AccountResponse struct {
	SecuritiesAccount SecuritiesAccount `json:"securitiesAccount"`
}

type SecuritiesAccount struct {
	AccountNumber           string            `json:"accountNumber,omitempty"`
	RoundTrips              int               `json:"roundTrips,omitempty"`
	IsDayTrader             bool              `json:"isDayTrader,omitempty"`
	IsClosingOnlyRestricted bool              `json:"isClosingOnlyRestricted,omitempty"`
	PfcbFlag                bool              `json:"pfcbFlag,omitempty"`
	Positions               []Position        `json:"positions,omitempty"`
	InitialBalances         BalancesInitial   `json:"initialBalances,omitempty"`
	CurrentBalances         BalancesCurrent   `json:"currentBalances,omitempty"`
	ProjectedBalances       BalancesProjected `json:"projectedBalances,omitempty"`
}

type Position struct {
	ShortQuantity                float64    `json:"shortQuantity,omitempty"`
	AveragePrice                 float64    `json:"averagePrice,omitempty"`
	CurrentDayProfitLoss         float64    `json:"currentDayProfitLoss,omitempty"`
	CurrentDayProfitLossPct      float64    `json:"currentDayProfitLossPercentage,omitempty"`
	LongQuantity                 float64    `json:"longQuantity,omitempty"`
	SettledLongQuantity          float64    `json:"settledLongQuantity,omitempty"`
	SettledShortQuantity         float64    `json:"settledShortQuantity,omitempty"`
	AgedQuantity                 float64    `json:"agedQuantity,omitempty"`
	Instrument                   Instrument `json:"instrument,omitempty"`
	MarketValue                  float64    `json:"marketValue,omitempty"`
	MaintenanceRequirement       float64    `json:"maintenanceRequirement,omitempty"`
	AverageLongPrice             float64    `json:"averageLongPrice,omitempty"`
	AverageShortPrice            float64    `json:"averageShortPrice,omitempty"`
	TaxLotAverageLongPrice       float64    `json:"taxLotAverageLongPrice,omitempty"`
	TaxLotAverageShortPrice      float64    `json:"taxLotAverageShortPrice,omitempty"`
	LongOpenProfitLoss           float64    `json:"longOpenProfitLoss,omitempty"`
	ShortOpenProfitLoss          float64    `json:"shortOpenProfitLoss,omitempty"`
	PreviousSessionLongQuantity  float64    `json:"previousSessionLongQuantity,omitempty"`
	PreviousSessionShortQuantity float64    `json:"previousSessionShortQuantity,omitempty"`
	CurrentDayCost               float64    `json:"currentDayCost,omitempty"`
}

type Instrument struct {
	AssetType    string  `json:"assetType,omitempty"`
	Cusip        string  `json:"cusip,omitempty"`
	Symbol       string  `json:"symbol,omitempty"`
	Description  string  `json:"description,omitempty"`
	InstrumentID int     `json:"instrumentId,omitempty"`
	NetChange    float64 `json:"netChange,omitempty"`
	Type         string  `json:"type,omitempty"`
}

type BalancesInitial struct {
	AccruedInterest                  float64 `json:"accruedInterest,omitempty"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade,omitempty"`
	BondValue                        float64 `json:"bondValue,omitempty"`
	BuyingPower                      float64 `json:"buyingPower,omitempty"`
	CashBalance                      float64 `json:"cashBalance,omitempty"`
	CashAvailableForTrading          float64 `json:"cashAvailableForTrading,omitempty"`
	CashReceipts                     float64 `json:"cashReceipts,omitempty"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower,omitempty"`
	DayTradingBuyingPowerCall        float64 `json:"dayTradingBuyingPowerCall,omitempty"`
	DayTradingEquityCall             float64 `json:"dayTradingEquityCall,omitempty"`
	Equity                           float64 `json:"equity,omitempty"`
	EquityPercentage                 float64 `json:"equityPercentage,omitempty"`
	LiquidationValue                 float64 `json:"liquidationValue,omitempty"`
	LongMarginValue                  float64 `json:"longMarginValue,omitempty"`
	LongOptionMarketValue            float64 `json:"longOptionMarketValue,omitempty"`
	LongStockValue                   float64 `json:"longStockValue,omitempty"`
	MaintenanceCall                  float64 `json:"maintenanceCall,omitempty"`
	MaintenanceRequirement           float64 `json:"maintenanceRequirement,omitempty"`
	Margin                           float64 `json:"margin,omitempty"`
	MarginEquity                     float64 `json:"marginEquity,omitempty"`
	MoneyMarketFund                  float64 `json:"moneyMarketFund,omitempty"`
	MutualFundValue                  float64 `json:"mutualFundValue,omitempty"`
	RegTCall                         float64 `json:"regTCall,omitempty"`
	ShortMarginValue                 float64 `json:"shortMarginValue,omitempty"`
	ShortOptionMarketValue           float64 `json:"shortOptionMarketValue,omitempty"`
	ShortStockValue                  float64 `json:"shortStockValue,omitempty"`
	TotalCash                        float64 `json:"totalCash,omitempty"`
	IsInCall                         bool    `json:"isInCall,omitempty"`
	UnsettledCash                    float64 `json:"unsettledCash,omitempty"`
	PendingDeposits                  float64 `json:"pendingDeposits,omitempty"`
	MarginBalance                    float64 `json:"marginBalance,omitempty"`
	ShortBalance                     float64 `json:"shortBalance,omitempty"`
	AccountValue                     float64 `json:"accountValue,omitempty"`
}

type BalancesCurrent struct {
	AvailableFunds                   float64 `json:"availableFunds,omitempty"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade,omitempty"`
	BuyingPower                      float64 `json:"buyingPower,omitempty"`
	BuyingPowerNonMarginableTrade    float64 `json:"buyingPowerNonMarginableTrade,omitempty"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower,omitempty"`
	DayTradingBuyingPowerCall        float64 `json:"dayTradingBuyingPowerCall,omitempty"`
	Equity                           float64 `json:"equity,omitempty"`
	EquityPercentage                 float64 `json:"equityPercentage,omitempty"`
	LongMarginValue                  float64 `json:"longMarginValue,omitempty"`
	MaintenanceCall                  float64 `json:"maintenanceCall,omitempty"`
	MaintenanceRequirement           float64 `json:"maintenanceRequirement,omitempty"`
	MarginBalance                    float64 `json:"marginBalance,omitempty"`
	RegTCall                         float64 `json:"regTCall,omitempty"`
	ShortBalance                     float64 `json:"shortBalance,omitempty"`
	ShortMarginValue                 float64 `json:"shortMarginValue,omitempty"`
	Sma                              float64 `json:"sma,omitempty"`
	IsInCall                         float64 `json:"isInCall,omitempty"`
	StockBuyingPower                 float64 `json:"stockBuyingPower,omitempty"`
	OptionBuyingPower                float64 `json:"optionBuyingPower,omitempty"`
}

type BalancesProjected = BalancesCurrent // Same structure

type AccountNumbers struct {
	AccountNumber string `json:"accountNumber"`
	HashValue     string `json:"hashValue"`
}

type AccountNumbersResponse []AccountNumbers

type Order struct {
	Session                  string          `json:"session,omitempty"`
	Duration                 string          `json:"duration,omitempty"`
	OrderType                string          `json:"orderType,omitempty"`
	CancelTime               time.Time       `json:"cancelTime,omitempty"`
	ComplexOrderStrategyType string          `json:"complexOrderStrategyType,omitempty"`
	Quantity                 float64         `json:"quantity,omitempty"`
	FilledQuantity           float64         `json:"filledQuantity,omitempty"`
	RemainingQuantity        float64         `json:"remainingQuantity,omitempty"`
	DestinationLinkName      string          `json:"destinationLinkName,omitempty"`
	ReleaseTime              time.Time       `json:"releaseTime,omitempty"`
	StopPrice                float64         `json:"stopPrice,omitempty"`
	StopPriceLinkBasis       string          `json:"stopPriceLinkBasis,omitempty"`
	StopPriceLinkType        string          `json:"stopPriceLinkType,omitempty"`
	StopPriceOffset          float64         `json:"stopPriceOffset,omitempty"`
	StopType                 string          `json:"stopType,omitempty"`
	PriceLinkBasis           string          `json:"priceLinkBasis,omitempty"`
	PriceLinkType            string          `json:"priceLinkType,omitempty"`
	Price                    float64         `json:"price,omitempty"`
	TaxLotMethod             string          `json:"taxLotMethod,omitempty"`
	OrderLegCollection       []OrderLeg      `json:"orderLegCollection,omitempty"`
	ActivationPrice          float64         `json:"activationPrice,omitempty"`
	SpecialInstruction       string          `json:"specialInstruction,omitempty"`
	OrderStrategyType        string          `json:"orderStrategyType,omitempty"`
	OrderID                  int64           `json:"orderId,omitempty"`
	Cancelable               bool            `json:"cancelable,omitempty"`
	Editable                 bool            `json:"editable,omitempty"`
	Status                   string          `json:"status,omitempty"`
	EnteredTime              time.Time       `json:"enteredTime,omitempty"`
	CloseTime                time.Time       `json:"closeTime,omitempty"`
	AccountNumber            int64           `json:"accountNumber,omitempty"`
	OrderActivityCollection  []OrderActivity `json:"orderActivityCollection,omitempty"`
	ReplacingOrderCollection []string        `json:"replacingOrderCollection,omitempty"`
	ChildOrderStrategies     []string        `json:"childOrderStrategies,omitempty"`
	StatusDescription        string          `json:"statusDescription,omitempty"`
}

type OrderLeg struct {
	OrderLegType   string     `json:"orderLegType,omitempty"`
	LegID          int64      `json:"legId,omitempty"`
	Instrument     Instrument `json:"instrument,omitempty"`
	Instruction    string     `json:"instruction,omitempty"`
	PositionEffect string     `json:"positionEffect,omitempty"`
	Quantity       float64    `json:"quantity,omitempty"`
	QuantityType   string     `json:"quantityType,omitempty"`
	DivCapGains    string     `json:"divCapGains,omitempty"`
	ToSymbol       string     `json:"toSymbol,omitempty"`
}

type OrderActivity struct {
	ActivityType           string         `json:"activityType,omitempty"`
	ExecutionType          string         `json:"executionType,omitempty"`
	Quantity               float64        `json:"quantity,omitempty"`
	OrderRemainingQuantity float64        `json:"orderRemainingQuantity,omitempty"`
	ExecutionLegs          []ExecutionLeg `json:"executionLegs,omitempty"`
}

type ExecutionLeg struct {
	LegID             int64     `json:"legId,omitempty"`
	Price             float64   `json:"price,omitempty"`
	Quantity          float64   `json:"quantity,omitempty"`
	MismarkedQuantity float64   `json:"mismarkedQuantity,omitempty"`
	InstrumentID      int64     `json:"instrumentId,omitempty"`
	Time              time.Time `json:"time,omitempty"`
}
