package egoscale

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (exo *Client) CreateEgressRule(rule SecurityGroupRule) (*AuthorizeSecurityGroupEgressResponse, error) {

	params := url.Values{}
	params.Set("securitygroupid", rule.SecurityGroupId)
	
        if rule.Cidr != "" {
            params.Set("cidrlist", rule.Cidr)
        } else if len(rule.UserSecurityGroupList) > 0 {
            usg,err := json.Marshal(rule.UserSecurityGroupList)
	    if err != nil {
	        	return nil, err
	    }
	    params.Set("usersecuritygrouplist", string(usg))
        } else {
            return nil, fmt.Errorf("No Egress rule CIDR or Security Group List provided")
        }

	params.Set("protocol", rule.Protocol)
        
	if rule.Protocol == "ICMP" {
		params.Set("icmpcode", fmt.Sprintf("%d", rule.IcmpCode))
		params.Set("icmptype", fmt.Sprintf("%d", rule.IcmpType))
	} else if rule.Protocol == "TCP" || rule.Protocol == "UDP" {
		params.Set("startport", fmt.Sprintf("%d", rule.Port))
		params.Set("endport", fmt.Sprintf("%d", rule.Port))
	} else {
		return nil, fmt.Errorf("Invalid Egress rule Protocol: %s", rule.Protocol)
	}

	resp, err := exo.Request("authorizeSecurityGroupEgress", params)
	if err != nil {
		return nil, err
	}

	var r AuthorizeSecurityGroupEgressResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (exo *Client) CreateIngressRule(rule SecurityGroupRule) (*AuthorizeSecurityGroupIngressResponse, error) {

	params := url.Values{}
	params.Set("securitygroupid", rule.SecurityGroupId)
        
        if rule.Cidr != "" {
            params.Set("cidrlist", rule.Cidr)
        } else if len(rule.UserSecurityGroupList) >0 {
            for i := 0; i < len(rule.UserSecurityGroupList); i++ {
                    params.Set(fmt.Sprintf("usersecuritygrouplist[%d].account", i), rule.UserSecurityGroupList[i].Account)
                    params.Set(fmt.Sprintf("usersecuritygrouplist[%d].group", i), rule.UserSecurityGroupList[i].Group)
                    
                }
        } else {
            return nil, fmt.Errorf("No Ingress rule CIDR or Security Group List provided")
        }

	params.Set("protocol", rule.Protocol)

	if rule.Protocol == "ICMP" {
		params.Set("icmpcode", fmt.Sprintf("%d", rule.IcmpCode))
		params.Set("icmptype", fmt.Sprintf("%d", rule.IcmpType))
	} else if rule.Protocol == "TCP" || rule.Protocol == "UDP" {
		params.Set("startport", fmt.Sprintf("%d", rule.Port))
		params.Set("endport", fmt.Sprintf("%d", rule.Port))
	} else {
		return nil, fmt.Errorf("Invalid Egress rule Protocol: %s", rule.Protocol)
	}

        fmt.Printf("## params: %+v\n", params)

	resp, err := exo.Request("authorizeSecurityGroupIngress", params)

	if err != nil {
		return nil, err
	}

	var r AuthorizeSecurityGroupIngressResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (exo *Client) CreateSecurityGroupWithRules(name string, ingress []SecurityGroupRule, egress []SecurityGroupRule) (*CreateSecurityGroupResponse, error) {

	params := url.Values{}
	params.Set("name", name)

	resp, err := exo.Request("createSecurityGroup", params)

	var r CreateSecurityGroupResponseWrapper
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	sgid := r.Wrapped.Id

	for _, erule := range egress {
		erule.SecurityGroupId = sgid
		_, err = exo.CreateEgressRule(erule)
		if err != nil {
			return nil, err
		}
	}

	for _, inrule := range ingress {
		inrule.SecurityGroupId = sgid
		_, err = exo.CreateIngressRule(inrule)
		if err != nil {
			return nil, err
		}
	}

	return &r.Wrapped, nil
}

func (exo *Client) DeleteSecurityGroup(name string) (error) {
	params := url.Values{}
	params.Set("name", name)

	resp, err := exo.Request("deleteSecurityGroup", params); if err != nil {
		return err
	}

	fmt.Printf("## response: %+v\n", resp)
	return nil
}
