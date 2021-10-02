package alertmanager

type Alert struct {
	Status      string
	Labels      map[string]string
	Annotations map[string]string
	StartsAt    string
	Fingerprint string
}

type WebhookPayload struct {
	Version  string
	Status   string
	Receiver string
	Alerts   []*Alert
}
