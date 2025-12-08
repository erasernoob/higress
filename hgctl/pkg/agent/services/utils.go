package services

import (
	"fmt"
	"net"
	"net/url"
)

func BuildAIProviderServiceBody(name, url string) map[string]interface{} {
	customBaseURL := fmt.Sprintf("%s/compatible-mode/v1", url)
	return map[string]interface{}{
		"type":     "openai",
		"name":     name,
		"tokens":   []string{},
		"version":  0,
		"protocol": "openai/v1",
		"tokenFailoverConfig": map[string]interface{}{
			"enabled": false,
		},
		"proxyName": "",
		"rawConfigs": map[string]interface{}{
			"openaiExtraCustomUrls": []string{},
			"openaiCustomUrl":       customBaseURL,
		},
	}
}

func BuildAIRouteServiceBody(name, _url string) map[string]interface{} {
	return map[string]interface{}{
		"name": fmt.Sprintf("%s-route", name),
		// "version": "627198", // It's unecessary to provide when create a new one
		"domains": []interface{}{},
		"pathPredicate": map[string]interface{}{
			"matchType":     "PRE",
			"matchValue":    "/",
			"caseSensitive": false,
		},
		"headerPredicates":   []interface{}{},
		"urlParamPredicates": []interface{}{},
		"upstreams": []interface{}{
			map[string]interface{}{
				"provider":     name,
				"weight":       100,
				"modelMapping": map[string]interface{}{},
			},
		},
		"modelPredicates": []interface{}{},
		"authConfig": map[string]interface{}{
			"enabled":                false,
			"allowedCredentialTypes": nil,
			"allowedConsumers":       []interface{}{},
		},
		"fallbackConfig": map[string]interface{}{
			"enabled":          false,
			"upstreams":        nil,
			"fallbackStrategy": nil,
			"responseCodes":    nil,
		},
	}
}

func BuildServiceBodyAndSrvName(name, urlStr string) (map[string]interface{}, string, error) {
	res, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}

	// add service source
	srvType := ""
	srvPort := ""

	if ip := net.ParseIP(res.Hostname()); ip == nil {
		srvType = "dns"
	} else {
		srvType = "static"
	}

	if res.Port() == "" && res.Scheme == "http" {
		srvPort = "80"
	} else if res.Port() == "" && res.Scheme == "https" {
		srvPort = "443"
	} else {
		srvPort = res.Port()
	}

	// e.g. agent-jarvis.static.8090
	targetSrvName := fmt.Sprintf("%s.%s:%s", name, srvType, srvPort)

	return map[string]interface{}{
		"domain":        res.Host,
		"type":          srvType,
		"port":          srvPort,
		"name":          name,
		"domainForEdit": res.Host,
		"protocol":      res.Scheme,
	}, targetSrvName, nil
}

func BuildAPIRouteBody(name, srv string) map[string]interface{} {
	return map[string]interface{}{
		"name": fmt.Sprintf("%s-route", name),
		"path": map[string]interface{}{
			"matchType":     "PRE",      // default is PREFIX
			"matchValue":    "/process", // default is "/process"
			"caseSensitive": true,
		},
		"authConfig": map[string]interface{}{
			"enabled": false,
		},
		"services": []map[string]interface{}{
			{
				"name": srv,
			},
		},
	}
}

func BuildAddHigressInstanceBody(name, addr, username, password string) map[string]interface{} {
	return map[string]interface{}{
		"gatewayName": name,
		"gatewayType": "HIGRESS",
		"higressConfig": map[string]interface{}{
			"address":  addr,
			"username": username,
			"password": password,
		},
	}
}

func BuildAPIProductBody(name, desc, typ string) map[string]interface{} {
	return map[string]interface{}{
		"name": name, "description": desc, "type": typ,
	}
}

func BuildRefModelAPIProductBody(gateway_id, product_id, target_route string) map[string]interface{} {
	return map[string]interface{}{
		"gatewayId":  gateway_id,
		"sourceType": "GATEWAY",
		"productId":  product_id,
		"higressRefConfig": map[string]interface{}{
			"modelRouteName":  target_route,
			"fromGatewayType": "HIGRESS",
		},
	}
}

func BuildRefMCPAPIProductBody(gateway_id, product_id, mcp_name string) map[string]interface{} {
	return map[string]interface{}{
		"gatewayId":  gateway_id,
		"sourceType": "GATEWAY",
		"productId":  product_id,
		"higressRefConfig": map[string]interface{}{
			"mcpServerName":   mcp_name,
			"fromGatewayType": "HIGRESS",
		},
	}
}
