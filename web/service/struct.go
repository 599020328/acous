package service

type BeginReq struct {
	SourceURL string `json:"source_url"`

	WorkChain string `json:"work_chain"`
}

type GetNeighborDelayReq struct {
	DataSize  int   `json:"data_size"`
	Round     int   `json:"round"`
	TimeStamp int64 `json:"time_stamp"`
}

type GetDelayResp struct {
	Code      int     `json:"code"`
	Message   string  `json:"message"`
	TimeDelay float64 `json:"time_delay"`
}

type RequestReq struct {
	SourceURL   string `json:"source_url"`
	DestNumber  string `json:"dest_number"`
	Round       int    `json:"round"`
	AntNumber   int    `json:"ant_number"`
	IsLastAnt   bool   `json:"is_last_ant"`
	IsLastRound bool   `json:"is_last_round"`

	Data          string `json:"data"`
	TimeStampNano int64  `json:"time_stamp_nano"`

	Path            string  `json:"path"`
	PathDelay       float64 `json:"path_delay"`
	PathDelayDetail string  `json:"path_delay_detail"`
}

type DestNodeReq struct {
	SourceURL       string  `json:"source_url"`
	DestNumber      string  `json:"dest_number"`
	Round           int     `json:"round"`
	AntNumber       int     `json:"ant_number"`
	IsLastAnt       bool    `json:"is_last_ant"`
	IsLastRound     bool    `json:"is_last_round"`
	Path            string  `json:"path"`
	PathDelay       float64 `json:"path_delay"`
	PathDelayDetail string  `json:"path_delay_detail"`

	IsSuccess bool `json:"is_success"`
}

type UpdateTauReq struct {
	Path     string `json:"path"`
	PathCost string `json:"path_cost"`
	Round    int    `json:"round"`

	Rho float64 `json:"rho"`
	Q   int     `json:"q"`
}

type TestReq struct {
	TimeStamp     int64  `json:"time_stamp"`
	TimeStampNano int64  `json:"time_stamp_nano"`
	NextURL       string `json:"next_url"`
}

type UpdateDataReq struct {
	DataDelta string	`json:"data_delta"`
}
