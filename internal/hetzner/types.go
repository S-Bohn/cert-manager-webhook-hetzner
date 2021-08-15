package hetzner

type Record struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	//Created  time.Time `json:"created"`
	//Modified time.Time `json:"modified"`
	ZoneID string `json:"zone_id"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	//TTL      uint64    `json:"ttl"`
}

type RecordInfo struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   uint64 `json:"ttl"`
}

type Zone struct {
	//Created         *time.Time `json:"created"`
	ID string `json:"id"`
	//IsSecondaryDNS  bool       `json:"is_secondary_dns"`
	//LegacyDNSHost   string     `json:"legacy_dns_host"`
	//LegacyNS        []string   `json:"legacy_ns"`
	//Modified        *time.Time `json:"modified"`
	Name string `json:"name"`
	//NS              []string   `json:"ns"`
	//Owner           string     `json:"owner"`
	//Permission      string     `json:"permission"`
	//Project         string     `json:"project"`
	//RecordsCount    uint64     `json:"records_count"`
	//Registrar       string     `json:"registrar"`
	//Status          string     `json:"status"`
	//TTL             uint64     `json:"ttl"`
	//TxtVerification struct {
	//	Name  string `json:"name"`
	//	Token string `json:"token"`
	//} `json:"txt_verification"`
	//Verified *time.Time `json:"verified"`
}

type createRecordRequest struct {
	ZoneID string `json:"zone_id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	TTL    uint64 `json:"ttl"`
}

type createRecordResponse struct {
	Record Record `json:"record"`
}

type getAllZonesResponse struct {
	Zones []Zone `json:"zones"`
	Meta  struct {
		Pagination struct {
			LastPage     uint32 `json:"last_page"`
			Page         uint32 `json:"page"`
			PerPage      uint32 `json:"per_page"`
			TotalEntries uint32 `json:"total_entries"`
		} `json:"pagination"`
	} `json:"meta"`
}

type getAllRecordsResponse struct {
	Records []Record `json:"records"`
}
