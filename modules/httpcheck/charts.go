package httpcheck

import (
	"github.com/netdata/go-orchestrator/module"
)

type (
	// Charts is an alias for module.Charts
	Charts = module.Charts
	// Chart is an alias for module.Chart
	Chart = module.Chart
	// Dims is an alias for module.Dims
	Dims = module.Dims
)

var charts = Charts{
	{
		ID:    "response_time",
		Title: "HTTP Response Time",
		Units: "ms",
		Fam:   "response",
		Ctx:   "httpcheck.responsetime",
		Dims: Dims{
			{ID: "time"},
		},
	},
	{
		ID:    "request_status",
		Title: "HTTP Check Status",
		Units: "boolean",
		Fam:   "status",
		Ctx:   "httpcheck.status",
		Dims: Dims{
			{ID: "success"},
			{ID: "no_connection", Name: "no connection"},
			{ID: "timeout"},
			{ID: "bad_content", Name: "bad content"},
			{ID: "bad_status", Name: "bad status"},
			//{ID: "dns_lookup_error", Name: "dns lookup error"},
			//{ID: "address_parse_error", Name: "address parse error"},
			//{ID: "redirect_error", Name: "redirect error"},
			//{ID: "body_read_error", Name: "body read error"},
		},
	},
}

var bodyLengthChart = Chart{
	ID:    "response_length",
	Title: "HTTP Response Body Length",
	Units: "characters",
	Fam:   "response",
	Ctx:   "httpcheck.responselength",
	Dims: Dims{
		{ID: "length"},
	},
}
