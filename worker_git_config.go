package brignext

type WorkerGitConfig struct {
	CloneURL       string `json:"cloneURL" bson:"cloneURL"`
	Commit         string `json:"commit" bson:"commit"`
	Ref            string `json:"ref" bson:"ref"`
	InitSubmodules bool   `json:"initSubmodules" bson:"initSubmodules"`
	// // TODO: We MUST encrypt this!
	// SSHKey  string `json:"sshKey" bson:"sshKey"`
	// SSHCert string `json:"sshCert" bson:"sshCert"`
}
