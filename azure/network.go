// Get networking metrics from Azure's Resource Manager.

package azure

import (
	"encoding/json"
	"strings"
)

// VPNConnection represents an Azure VPN connection object
type VPNConnection struct {
	ID            string `json:"id"`
	Location      string `json:"location"`
	Name          string `json:"name"`
	ResourceGroup string `json:"name,omitempty"`
	Properties    struct {
		Status                  string  `json:"connectionStatus"`
		EgressBytesTransferred  float64 `json:"egressBytesTransferred"`
		IngressBytesTransferred float64 `json:"ingressBytesTransferred"`
	}
}

// VPNConnectionsListResult is used to unmarshal the data from Azure when getting a list of VPN connections
type VPNConnectionsListResult struct {
	Connections []VPNConnection `json:"value"`
}

// FindVPNConnections returns a slice of all VPN connection objects for the current subscription
func (c *Client) FindVPNConnections() ([]VPNConnection, error) {
	u := c.buildURL("", "Microsoft.Network/connections", "2015-06-15")
	resp, err := c.get(u.String())
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	var conns VPNConnectionsListResult
	if err = decoder.Decode(&conns); err != nil {
		return nil, err
	}

	for idx, conn := range conns.Connections {
		conn.Name = strings.Split(conn.ID, "/")[8]
		conn.ResourceGroup = strings.Split(conn.ID, "/")[4]
		conns.Connections[idx] = conn
	}
	return conns.Connections, nil
}

// FindVPNConnection gets the given VPN connection by namd and resource group if found
func (c *Client) FindVPNConnection(group, name string) (VPNConnection, error) {
	u := c.buildURL(group, "Microsoft.Network/connections/"+name, "2015-06-15")
	resp, err := c.get(u.String())
	if err != nil {
		return VPNConnection{}, err
	}
	defer resp.Body.Close()
	var conn VPNConnection
	if err = json.NewDecoder(resp.Body).Decode(&conn); err != nil {
		return VPNConnection{}, err
	}
	return conn, nil
}
