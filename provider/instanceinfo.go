package provider

type InstanceInfo struct {
	Name          string `json:"name,omitempty"`
	State         string `json:"state,omitempty"`
	InitialState  string `json:"initialstate,omitempty"`
	ExternalIp    string `json:"externalip,omitempty"`
	SSHPort       string `json:"sshport,omitempty"`
	Ip            string `json:"ip,omitempty"`
	User          string `json:"user,omitempty"`
	PackIdExist   bool   `json:"packidexist,omitempty"`
	SystemIdExist bool   `json:"systemidexist,omitempty"`
}
