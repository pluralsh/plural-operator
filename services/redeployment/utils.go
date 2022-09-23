package redeployment

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func RedeployLabelSelector() (labels.Selector, error) {
	req, err := labels.NewRequirement(RedeployLabel, selection.Equals, []string{"true"})
	if err != nil {
		return nil, fmt.Errorf("failed to build label selector: %w", err)
	}
	return labels.Parse(req.String())
}
