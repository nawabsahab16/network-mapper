package cloudclient

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/otterize/intents-operator/src/shared/otterizecloud/otterizecloudclient"
	"github.com/sirupsen/logrus"
)

type CloudClient interface {
	ReportDiscoveredIntents(ctx context.Context, intents []*DiscoveredIntentInput) error
	ReportComponentStatus(ctx context.Context, component ComponentType) error
}

type CloudClientImpl struct {
	client graphql.Client
}

func NewClient(ctx context.Context) (CloudClient, bool, error) {
	client, ok, err := otterizecloudclient.NewClient(ctx)
	if !ok {
		return nil, false, nil
	} else if err != nil {
		return nil, true, err
	}

	return &CloudClientImpl{client: client}, true, nil
}

func (c *CloudClientImpl) ReportDiscoveredIntents(ctx context.Context, intents []*DiscoveredIntentInput) error {
	logrus.Info("Uploading intents to cloud, count: ", len(intents))

	_, err := ReportDiscoveredIntents(ctx, c.client, intents)
	if err != nil {
		return err
	}

	return nil
}

func (c *CloudClientImpl) ReportComponentStatus(ctx context.Context, component ComponentType) error {
	logrus.Info("Uploading component status to cloud")

	_, err := ReportComponentStatus(ctx, c.client, component)
	if err != nil {
		return err
	}
	return nil
}
