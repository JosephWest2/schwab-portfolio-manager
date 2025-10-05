package marketData


type QuoteResponse map[string]Instrument

type Instrument struct {
	AssetMainType string       `json:"assetMainType"`
	AssetSubType  string       `json:"assetSubType,omitempty"`
	Symbol        string       `json:"symbol"`
	QuoteType     string       `json:"quoteType,omitempty"`
	Realtime      bool         `json:"realtime"`
	SSID          int64        `json:"ssid"`
	Reference     Reference    `json:"reference"`
	Quote         Quote        `json:"quote"`
	Regular       *Regular     `json:"regular,omitempty"`
	Fundamental   *Fundamental `json:"fundamental,omitempty"`
}

type Reference struct {
	CUSIP                string  `json:"cusip,omitempty"`
	Description          string  `json:"description"`
	Exchange             string  `json:"exchange"`
	ExchangeName         string  `json:"exchangeName"`
	OtcMarketTier        string  `json:"otcMarketTier,omitempty"`
	ContractType         string  `json:"contractType,omitempty"`
	DaysToExpiration     int     `json:"daysToExpiration,omitempty"`
	ExpirationDay        int     `json:"expirationDay,omitempty"`
	ExpirationMonth      int     `json:"expirationMonth,omitempty"`
	ExpirationYear       int     `json:"expirationYear,omitempty"`
	IsPennyPilot         bool    `json:"isPennyPilot,omitempty"`
	LastTradingDay       int64   `json:"lastTradingDay,omitempty"`
	Multiplier           int     `json:"multiplier,omitempty"`
	SettlementType       string  `json:"settlementType,omitempty"`
	StrikePrice          float64 `json:"strikePrice,omitempty"`
	Underlying           string  `json:"underlying,omitempty"`
	UvExpirationType     string  `json:"uvExpirationType,omitempty"`
	FutureActiveSymbol   string  `json:"futureActiveSymbol,omitempty"`
	FutureExpirationDate int64   `json:"futureExpirationDate,omitempty"`
}

type Quote struct {
	Week52High          float64 `json:"52WeekHigh,omitempty"`
	Week52Low           float64 `json:"52WeekLow,omitempty"`
	AskMICId            string  `json:"askMICId,omitempty"`
	AskPrice            float64 `json:"askPrice,omitempty"`
	AskSize             int64   `json:"askSize,omitempty"`
	AskTime             int64   `json:"askTime,omitempty"`
	BidMICId            string  `json:"bidMICId,omitempty"`
	BidPrice            float64 `json:"bidPrice,omitempty"`
	BidSize             int64   `json:"bidSize,omitempty"`
	BidTime             int64   `json:"bidTime,omitempty"`
	ClosePrice          float64 `json:"closePrice,omitempty"`
	HighPrice           float64 `json:"highPrice,omitempty"`
	LastMICId           string  `json:"lastMICId,omitempty"`
	LastPrice           float64 `json:"lastPrice,omitempty"`
	LastSize            int64   `json:"lastSize,omitempty"`
	LowPrice            float64 `json:"lowPrice,omitempty"`
	Mark                float64 `json:"mark,omitempty"`
	MarkChange          float64 `json:"markChange,omitempty"`
	MarkPercentChange   float64 `json:"markPercentChange,omitempty"`
	NetChange           float64 `json:"netChange,omitempty"`
	NetPercentChange    float64 `json:"netPercentChange,omitempty"`
	OpenPrice           float64 `json:"openPrice,omitempty"`
	QuoteTime           int64   `json:"quoteTime,omitempty"`
	SecurityStatus      string  `json:"securityStatus,omitempty"`
	TotalVolume         int64   `json:"totalVolume,omitempty"`
	TradeTime           int64   `json:"tradeTime,omitempty"`
	Volatility          float64 `json:"volatility,omitempty"`
	NAV                 float64 `json:"nAV,omitempty"`
	Delta               float64 `json:"delta,omitempty"`
	Gamma               float64 `json:"gamma,omitempty"`
	ImpliedYield        float64 `json:"impliedYield,omitempty"`
	IndAskPrice         float64 `json:"indAskPrice,omitempty"`
	IndBidPrice         float64 `json:"indBidPrice,omitempty"`
	IndQuoteTime        int64   `json:"indQuoteTime,omitempty"`
	MoneyIntrinsicValue float64 `json:"moneyIntrinsicValue,omitempty"`
	OpenInterest        int64   `json:"openInterest,omitempty"`
	Rho                 float64 `json:"rho,omitempty"`
	TheoreticalValue    float64 `json:"theoreticalOptionValue,omitempty"`
	Theta               float64 `json:"theta,omitempty"`
	TimeValue           float64 `json:"timeValue,omitempty"`
	UnderlyingPrice     float64 `json:"underlyingPrice,omitempty"`
	Vega                float64 `json:"vega,omitempty"`
}

type Regular struct {
	RegularMarketLastPrice     float64 `json:"regularMarketLastPrice"`
	RegularMarketLastSize      int64   `json:"regularMarketLastSize"`
	RegularMarketNetChange     float64 `json:"regularMarketNetChange"`
	RegularMarketPercentChange float64 `json:"regularMarketPercentChange"`
	RegularMarketTradeTime     int64   `json:"regularMarketTradeTime"`
}

type Fundamental struct {
	Avg10DaysVolume    int64   `json:"avg10DaysVolume"`
	Avg1YearVolume     int64   `json:"avg1YearVolume"`
	DeclarationDate    string  `json:"declarationDate,omitempty"`
	DivAmount          float64 `json:"divAmount"`
	DivExDate          string  `json:"divExDate,omitempty"`
	DivFreq            int     `json:"divFreq"`
	DivPayAmount       float64 `json:"divPayAmount"`
	DivPayDate         string  `json:"divPayDate,omitempty"`
	DivYield           float64 `json:"divYield"`
	EPS                float64 `json:"eps"`
	FundLeverageFactor float64 `json:"fundLeverageFactor"`
	NextDivExDate      string  `json:"nextDivExDate,omitempty"`
	NextDivPayDate     string  `json:"nextDivPayDate,omitempty"`
	PERatio            float64 `json:"peRatio"`
	FundStrategy       string  `json:"fundStrategy,omitempty"`
}
