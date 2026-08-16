package zabbix

import (
	"fmt"
	"strings"
)

type ResourceInfo struct {
	Name             string
	Aliases          []string
	APIPrefix        string
	IDProperty       string
	PluralIDProperty string
	NameProperty     string
	Description      string
	SearchFields     []string
}

var ResourceMap = map[string]ResourceInfo{
	"host": {
		Name:             "host",
		Aliases:          []string{"h", "hosts"},
		APIPrefix:        "host",
		IDProperty:       "hostid",
		PluralIDProperty: "hostids",
		NameProperty:     "name",
		Description:      "Zabbix monitored hosts",
		SearchFields:     []string{"host", "name"},
	},
	"inventory": {
		Name:             "inventory",
		Aliases:          []string{"inv", "host-inventory", "inventories"},
		APIPrefix:        "host",
		IDProperty:       "hostid",
		PluralIDProperty: "hostids",
		NameProperty:     "name",
		Description:      "Zabbix host asset and hardware inventory",
		SearchFields:     []string{"host", "name", "type", "vendor", "model", "macaddress_a", "serialno_a", "os"},
	},
	"problem": {
		Name:             "problem",
		Aliases:          []string{"p", "problems"},
		APIPrefix:        "problem",
		IDProperty:       "eventid",
		PluralIDProperty: "eventids",
		NameProperty:     "name",
		Description:      "Active Zabbix problems",
		SearchFields:     []string{"name"},
	},
	"item": {
		Name:             "item",
		Aliases:          []string{"i", "items"},
		APIPrefix:        "item",
		IDProperty:       "itemid",
		PluralIDProperty: "itemids",
		NameProperty:     "name",
		Description:      "Host items/metrics",
		SearchFields:     []string{"name", "key_"},
	},
	"trigger": {
		Name:             "trigger",
		Aliases:          []string{"t", "triggers"},
		APIPrefix:        "trigger",
		IDProperty:       "triggerid",
		PluralIDProperty: "triggerids",
		NameProperty:     "description",
		Description:      "Trigger evaluation rules",
		SearchFields:     []string{"description", "comments"},
	},
	"template": {
		Name:             "template",
		Aliases:          []string{"tmpl", "templates"},
		APIPrefix:        "template",
		IDProperty:       "templateid",
		PluralIDProperty: "templateids",
		NameProperty:     "name",
		Description:      "Configuration templates",
		SearchFields:     []string{"host", "name"},
	},
	"hostgroup": {
		Name:             "hostgroup",
		Aliases:          []string{"hg", "hostgroups"},
		APIPrefix:        "hostgroup",
		IDProperty:       "groupid",
		PluralIDProperty: "groupids",
		NameProperty:     "name",
		Description:      "Host groups",
		SearchFields:     []string{"name"},
	},
	"maintenance": {
		Name:             "maintenance",
		Aliases:          []string{"maint", "maintenances"},
		APIPrefix:        "maintenance",
		IDProperty:       "maintenanceid",
		PluralIDProperty: "maintenanceids",
		NameProperty:     "name",
		Description:      "Maintenance windows",
		SearchFields:     []string{"name", "description"},
	},
	"event": {
		Name:             "event",
		Aliases:          []string{"ev", "events"},
		APIPrefix:        "event",
		IDProperty:       "eventid",
		PluralIDProperty: "eventids",
		NameProperty:     "name",
		Description:      "System events",
		SearchFields:     []string{"name"},
	},
	"user": {
		Name:             "user",
		Aliases:          []string{"users"},
		APIPrefix:        "user",
		IDProperty:       "userid",
		PluralIDProperty: "userids",
		NameProperty:     "username",
		Description:      "Zabbix user accounts",
		SearchFields:     []string{"username", "name", "surname"},
	},
	"service": {
		Name:             "service",
		Aliases:          []string{"services"},
		APIPrefix:        "service",
		IDProperty:       "serviceid",
		PluralIDProperty: "serviceids",
		NameProperty:     "name",
		Description:      "IT Service monitoring",
		SearchFields:     []string{"name", "description"},
	},
	"sla": {
		Name:             "sla",
		Aliases:          []string{"slas"},
		APIPrefix:        "sla",
		IDProperty:       "slaid",
		PluralIDProperty: "slaids",
		NameProperty:     "name",
		Description:      "Service Level Agreements",
		SearchFields:     []string{"name", "description"},
	},
	"dashboard": {
		Name:             "dashboard",
		Aliases:          []string{"dashboards"},
		APIPrefix:        "dashboard",
		IDProperty:       "dashboardid",
		PluralIDProperty: "dashboardids",
		NameProperty:     "name",
		Description:      "Zabbix dashboards",
		SearchFields:     []string{"name"},
	},
	"history": {
		Name:             "history",
		Aliases:          []string{"metric", "metrics", "telemetry"},
		APIPrefix:        "history",
		IDProperty:       "itemid",
		PluralIDProperty: "itemids",
		NameProperty:     "name",
		Description:      "Historical metric telemetry values",
		SearchFields:     []string{},
	},
}

func ResolveResource(name string) (ResourceInfo, error) {
	lower := strings.ToLower(strings.TrimSpace(name))
	for _, info := range ResourceMap {
		if info.Name == lower {
			return info, nil
		}
		for _, alias := range info.Aliases {
			if alias == lower {
				return info, nil
			}
		}
	}
	return ResourceInfo{}, fmt.Errorf("unknown resource %q", name)
}
