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

//func (info InstanceInfo) Connect() (*exec2.SSHClient, *cmd.XbeeError) {
//	return exec2.Connect(info.ExternalIp, info.SSHPort, info.User)
//}
//
//func (info InstanceInfo) Enter(ctx context.Context, user string) *cmd.XbeeError {
//	args := []string{"-i", newfs.NewRsaGen("").RootKeyPEM().String(),
//		"-p", info.SSHPort,
//		"-o", "StrictHostKeyChecking=no"}
//	args = append(args, fmt.Sprintf("%s@%s", user, info.ExternalIp))
//	aCmd := exec.CommandContext(ctx, "ssh", args...)
//	aCmd.Stdout = os.Stdout
//	aCmd.Stderr = os.Stderr
//	aCmd.Stdin = os.Stdin
//	if err := aCmd.Run(); err != nil {
//		return cmd.Error("command [%s] failed: %v", aCmd.String(), err)
//	}
//	return nil
//}
